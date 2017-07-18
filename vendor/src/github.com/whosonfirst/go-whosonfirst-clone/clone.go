package clone

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	_ "fmt"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pool"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type WOFClone struct {
	Source         string
	Dest           string
	Success        int64
	Error          int64
	Skipped        int64
	Scheduled      int64
	Completed      int64
	MaxFilehandles int64
	Filehandles    int64
	MaxRetries     float64 // max percentage of errors over scheduled
	Failed         []string
	Logger         *log.WOFLogger
	client         *http.Client
	retries        *pool.LIFOPool
	writesync      *sync.WaitGroup
	timer          time.Time
	done           chan bool
	throttle       chan bool
}

func NewWOFClone(source string, dest string, procs int, logger *log.WOFLogger) (*WOFClone, error) {

	// https://golang.org/src/net/http/filetransport.go

	u, err := url.Parse(source)

	if err != nil {
		return nil, err
	}

	var cl *http.Client

	if u.Scheme == "file" {

		root := u.Path

		if !strings.HasSuffix(root, "/") {
			root = root + "/"
		}

		/*
			Pay attention to what's going here. Absent tweaking the URL to
			fetch in the 'Fetch' method the following will not work. In
			order to make this working *without* tweaking the URL you would
			need to specifiy the root as '/' which just seems like a bad
			idea. The fear of blindly opening up the root level directory on
			the file system in this context may seem a bit premature (not to
			mention silly) but measure twice and all that good stuff...
			See also: https://code.google.com/p/go/issues/detail?id=2113
			(20160112/thisisaaronland)
		*/

		t := &http.Transport{}
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir(root)))

		cl = &http.Client{Transport: t}
	} else {
		cl = &http.Client{}
	}

	runtime.GOMAXPROCS(procs)

	count := 1000 // anything more and the operating system's "too many open filehandles"
	// triggers get hit (20161229/thisisaaronland)

	throttle := make(chan bool, count)

	for i := 0; i < count; i++ {
		throttle <- true
	}

	retries := pool.NewLIFOPool()

	/*

		This gets triggered in the 'Process' function to ensure that
		we don't exit out of 'CloneMetaFile' before all the goroutines
		to write new files to disk actually finish ... you know, writing
		to disk (20160606/thisisaaronland)
	*/

	writesync := new(sync.WaitGroup)

	ch := make(chan bool)

	c := WOFClone{
		Success:        0,
		Error:          0,
		Skipped:        0,
		Filehandles:    0,
		MaxFilehandles: 512,
		Source:         source,
		Dest:           dest,
		Logger:         logger,
		MaxRetries:     25.0, // maybe allow this to be user-defined ?
		client:         cl,
		writesync:      writesync,
		retries:        retries,
		timer:          time.Now(),
		done:           ch,
		throttle:       throttle,
	}

	go func(c *WOFClone) {

		for {
			select {

			case <-c.done:
				break
			case <-time.After(1 * time.Second):
				c.Status()
			}
		}
	}(&c)

	return &c, nil
}

func (c *WOFClone) CloneMetaFile(file string, skip_existing bool, force_updates bool) error {

	abs_path, _ := filepath.Abs(file)

	reader, read_err := csv.NewDictReaderFromPath(abs_path)

	if read_err != nil {
		c.Logger.Error("Failed to read %s, because %v", abs_path, read_err)
		return read_err
	}

	wg := new(sync.WaitGroup)

	c.timer = time.Now()

	for {

		row, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		rel_path, ok := row["path"]

		if !ok {
			continue
		}

		ensure_changes := true
		has_changes := true
		carry_on := false

		remote := c.Source + rel_path
		local := path.Join(c.Dest, rel_path)

		_, err = os.Stat(local)

		if !os.IsNotExist(err) {

			if force_updates {

				c.Logger.Debug("%s already but we are forcing updates", local)
			} else if skip_existing {

				c.Logger.Debug("%s already exists and we are skipping things that exist", local)
				carry_on = true

			} else {

				file_hash, ok := row["file_hash"]

				t1 := time.Now()

				if ok {
					c.Logger.Debug("comparing hardcoded hash (%s) for %s", file_hash, local)
					has_changes, _ = c.HasHashChanged(file_hash, remote)
				} else {
					has_changes, _ = c.HasChanged(local, remote)
				}

				if !has_changes {
					c.Logger.Info("no changes to %s", local)
					carry_on = true
				}

				t2 := time.Since(t1)

				c.Logger.Debug("time to determine whether %s has changed (%t), %v", local, has_changes, t2)
			}

			if carry_on {

				atomic.AddInt64(&c.Scheduled, 1)
				atomic.AddInt64(&c.Completed, 1)
				atomic.AddInt64(&c.Skipped, 1)
				continue
			}

			ensure_changes = false
		}

		<-c.throttle

		wg.Add(1)
		atomic.AddInt64(&c.Scheduled, 1)

		go func(c *WOFClone, rel_path string, ensure_changes bool) {

			c.EnsureFilehandles()

			defer func() {
				c.throttle <- true
				wg.Done()
			}()

			t1 := time.Now()
			cl_err := c.ClonePath(rel_path, ensure_changes)
			t2 := time.Since(t1)

			c.Logger.Debug("time to process %s : %v", rel_path, t2)

			if cl_err != nil {
				atomic.AddInt64(&c.Error, 1)
				c.retries.Push(&pool.PoolString{String: rel_path})
			} else {
				atomic.AddInt64(&c.Success, 1)
			}

			atomic.AddInt64(&c.Completed, 1)

		}(c, rel_path, ensure_changes)
	}

	wg.Wait()

	c.writesync.Wait()

	ok := c.ProcessRetries()

	if !ok {
		c.Logger.Warning("failed to process retries")
		return errors.New("One of file failed to be cloned")
	}

	c.writesync.Wait()

	c.done <- true
	return nil
}

func (c *WOFClone) ProcessRetries() bool {

	to_retry := c.retries.Length()

	if to_retry > 0 {

		scheduled := atomic.LoadInt64(&c.Scheduled)
		scheduled_f := float64(scheduled)

		retry_f := float64(to_retry)

		pct := (retry_f / scheduled_f) * 100.0

		if pct > c.MaxRetries {
			c.Logger.Warning("E_EXCESSIVE_ERRORS, %f percent of scheduled processes failed thus undermining our faith that they will work now...", pct)
			return false
		}

		c.Logger.Info("There are %d failed requests that will now be retried", to_retry)

		wg := new(sync.WaitGroup)

		for c.retries.Length() > 0 {

			r, ok := c.retries.Pop()

			if !ok {
				c.Logger.Error("failed to pop retries because... computers?")
				break
			}

			if r == nil {
				c.Logger.Error("why is retry (pool) item nil?")
				break
			}

			rel_path := r.StringValue()

			<-c.throttle

			atomic.AddInt64(&c.Scheduled, 1)
			wg.Add(1)

			go func(c *WOFClone, rel_path string) {

				defer func() {
					wg.Done()
					c.throttle <- true
				}()

				ensure_changes := true

				t1 := time.Now()

				cl_err := c.ClonePath(rel_path, ensure_changes)

				t2 := time.Since(t1)

				c.Logger.Debug("time to retry clone %s : %v\n", rel_path, t2)

				if cl_err != nil {
					atomic.AddInt64(&c.Error, 1)
				} else {
					atomic.AddInt64(&c.Error, -1)
				}

				atomic.AddInt64(&c.Completed, 1)

			}(c, rel_path)
		}

		wg.Wait()
	}

	return true
}

func (c *WOFClone) ClonePath(rel_path string, ensure_changes bool) error {

	remote := c.Source + rel_path
	local := path.Join(c.Dest, rel_path)

	_, err := os.Stat(local)

	if !os.IsNotExist(err) && ensure_changes {

		change, _ := c.HasChanged(local, remote)

		if !change {

			c.Logger.Debug("%s has not changed so skipping", local)
			atomic.AddInt64(&c.Skipped, 1)
			return nil
		}

	}

	process_err := c.Process(remote, local)

	if process_err != nil {
		return process_err
	}

	return nil
}

// don't return true if there's a problem - move that logic up above

func (c *WOFClone) HasChanged(local string, remote string) (bool, error) {

	change := true

	// OPEN FH

	atomic.AddInt64(&c.Filehandles, 1)

	defer func() {
		atomic.AddInt64(&c.Filehandles, -1)
	}()

	// we used to do the following with a helper function in go-whosonfirst-utils
	// but that package has gotten unweildy and out of control - I am thinking about
	// a generic WOF "hashing" package but that started turning in to quicksand so
	// in the interest of just removing go-whosonfirst-utils as a dependency we're
	// going to do it the old-skool way by hand, for now (20170718/thisisaaronland)

	body, err := ioutil.ReadFile(local)

	if err != nil {
		return false, err
	}

	enc := md5.Sum(body)
	local_hash := hex.EncodeToString(enc[:])

	// see notes above

	/*
		hash, err := hash.NewMD5Hash()

		if err != nil {
		   return false, err
		}

		local_hash, err := hash.HashFile(local)

		if err != nil {
		   return false, err
		}
	*/

	if err != nil {
		c.Logger.Error("Failed to hash %s, becase %v", local, err)

		c.SetMaxFilehandles()
		return change, err
	}

	return c.HasHashChanged(local_hash, remote)
}

func (c *WOFClone) HasHashChanged(local_hash string, remote string) (bool, error) {

	change := true

	rsp, err := c.Fetch("HEAD", remote)

	if err != nil {
		return change, err
	}

	defer func() {
		rsp.Body.Close()
	}()

	etag := rsp.Header.Get("Etag")
	remote_hash := strings.Replace(etag, "\"", "", -1)

	if local_hash == remote_hash {
		change = false
	}

	return change, nil
}

func (c *WOFClone) Process(remote string, local string) error {

	c.Logger.Debug("fetch %s and store in %s", remote, local)

	local_root := path.Dir(local)

	_, err := os.Stat(local_root)

	if os.IsNotExist(err) {
		c.Logger.Debug("create %s", local_root)
		os.MkdirAll(local_root, 0755)
	}

	t1 := time.Now()

	rsp, fetch_err := c.Fetch("GET", remote)

	t2 := time.Since(t1)

	c.Logger.Debug("time to fetch %s: %v", remote, t2)

	if fetch_err != nil {
		return fetch_err
	}

	defer func() {
		rsp.Body.Close()
	}()

	contents, read_err := ioutil.ReadAll(rsp.Body)

	if read_err != nil {
		c.Logger.Error("failed to read body for %s, because %v", remote, read_err)
		return read_err
	}

	c.writesync.Add(1) // See notes above in 'NewWOFClone'

	go func(writesync *sync.WaitGroup, local string, contents []byte) error {

		defer writesync.Done()

		// OPEN FH

		atomic.AddInt64(&c.Filehandles, 1)

		write_err := ioutil.WriteFile(local, contents, 0644)

		atomic.AddInt64(&c.Filehandles, -1)

		if write_err != nil {

			c.Logger.Error("Failed to write %s, because %v", local, write_err)
			c.SetMaxFilehandles()

			atomic.AddInt64(&c.Success, -1)
			atomic.AddInt64(&c.Error, 1)

			return write_err
		}

		c.Logger.Debug("Wrote %s to disk", local)
		return nil

	}(c.writesync, local, contents)

	return nil
}

func (c *WOFClone) Fetch(method string, remote string) (*http.Response, error) {

	/*
	  See notes in NewWOFClone for details on what's going on here. Given that
	  we are already testing whether c.Source is a file URI parsing remote here
	  is probably a bit of a waste. We will live with the cost for now and optimize
	  as necessary... (20160112/thisisaaronland)
	*/

	u, _ := url.Parse(remote)

	if u.Scheme == "file" {
		remote = strings.Replace(remote, c.Source, "", 1)
		remote = "file:///" + remote
		c.Logger.Debug("remote is now %s", remote)
	}

	c.Logger.Debug("%s %s", method, remote)

	req, _ := http.NewRequest(method, remote, nil)
	req.Close = true

	// OPEN FH

	atomic.AddInt64(&c.Filehandles, 1)

	rsp, err := c.client.Do(req)

	atomic.AddInt64(&c.Filehandles, -1)

	if err != nil {

		c.Logger.Error("Failed to %s %s, because %v", method, remote, err)

		if u.Scheme == "file" {
			c.SetMaxFilehandles()
		}

		return nil, err
	}

	// Notice how we are not closing rsp.Body - that's because we are passing
	// it (rsp) back up the stack

	// See also: https://github.com/whosonfirst/go-whosonfirst-clone/issues/6

	expected := 200

	if rsp.StatusCode != expected {

		rsp.Body.Close()

		c.Logger.Error("Failed to %s %s, because we expected %d from source and got '%s' instead", method, remote, expected, rsp.Status)

		if u.Scheme == "file" {
			c.SetMaxFilehandles()
		}

		return nil, errors.New(rsp.Status)
	}

	return rsp, nil
}

func (c *WOFClone) SetMaxFilehandles() {

	// c.fh.Lock()
	// defer c.fh.Unlock()

	count := atomic.LoadInt64(&c.Filehandles)
	max := int64(float64(count) * 0.75)

	atomic.StoreInt64(&c.MaxFilehandles, max)

	c.Logger.Info("Set max filehandles to %d (triggered at %d)", max, count)
}

func (c *WOFClone) EnsureFilehandles() {

	current := atomic.LoadInt64(&c.Filehandles)
	max := atomic.LoadInt64(&c.MaxFilehandles)

	if max == 0 || current < max {
		return
	}

	select {

	case <-time.After(100 * time.Millisecond):

		// c.fh.Lock()

		current := atomic.LoadInt64(&c.Filehandles)
		max := atomic.LoadInt64(&c.MaxFilehandles)

		// c.fh.Unlock()

		if max == 0 || current < max {
			// c.Logger.Info("max file handles lock removed max: %d current: %d", max, current)
			break
		} else {
			// c.Logger.Info("max file handles in effect max: %d current: %d", max, current)
		}
	}
}

func (c *WOFClone) Status() {

	t2 := time.Since(c.timer)

	scheduled := atomic.LoadInt64(&c.Scheduled)
	completed := atomic.LoadInt64(&c.Completed)
	success := atomic.LoadInt64(&c.Success)
	error := atomic.LoadInt64(&c.Error)
	skipped := atomic.LoadInt64(&c.Skipped)

	current_fh := atomic.LoadInt64(&c.Filehandles)
	max_fh := atomic.LoadInt64(&c.MaxFilehandles)

	c.Logger.Info("scheduled: %d completed: %d success: %d error: %d skipped: %d to retry: %d goroutines: %d filehandles: %d/%d time: %v",
		scheduled, completed, success, error, skipped, c.retries.Length(), runtime.NumGoroutine(), current_fh, max_fh, t2)

	// https://deferpanic.com/blog/understanding-golang-memory-usage/
	// https://golang.org/pkg/runtime/#MemStats

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	c.Logger.Debug("memstats: alloc: %d total alloc: %d heap alloc: %d heap size: %d", mem.Alloc, mem.TotalAlloc, mem.HeapAlloc, mem.HeapSys)
}
