package desktop

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const bufferSize = 32 * 1024

type scan struct {
	e    [][]*Entry
	errs chan error
	in   chan *scanEntry
	sync.Mutex
	sync.WaitGroup
}

type scanEntry struct {
	i int
	f *os.File
}

// Scan non-recursively scans provided directories for desktop entry files and
// parses them. A slice of parsed entries is returned for each directory.
func Scan(dirs []string) ([][]*Entry, error) {
	s := &scan{e: make([][]*Entry, len(dirs)), errs: make(chan error), in: make(chan *scanEntry)}

	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
		go scanner(s)
	}

	for i, dir := range dirs {
		i, dir := i, dir

		s.Add(1)
		go scanDir(i, dir, s)
	}

	done := make(chan bool, 1)
	go func() {
		s.Wait()
		close(s.in)

		done <- true
	}()

	select {
	case err := <-s.errs:
		return nil, err
	case <-done:
		return s.e, nil
	}
}

func scanner(s *scan) {
	var (
		buf       = make([]byte, bufferSize)
		scanEntry *scanEntry
		entry     *Entry
		err       error
	)

	for scanEntry = range s.in {
		entry, err = Parse(scanEntry.f, buf)
		scanEntry.f.Close()
		if err != nil {
			s.errs <- err
			s.Done()
			return
		} else if entry == nil {
			s.Done()
			continue
		}

		s.Lock()
		s.e[scanEntry.i] = append(s.e[scanEntry.i], entry)
		s.Unlock()

		s.Done()
	}
}

func scanFile(i int, dir string, e os.DirEntry, s *scan) {
	if e == nil || e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".desktop") {
		s.Done()
		return
	}

	f, err := os.OpenFile(filepath.Join(dir, e.Name()), os.O_RDONLY, 0644)
	if os.IsNotExist(err) {
		s.Done()
		return
	} else if err != nil {

		s.errs <- err
		s.Done()
		return
	}

	s.in <- &scanEntry{i: i, f: f}
}

func scanDir(i int, dir string, s *scan) {
	defer s.Done()

	dirEntries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Fatal(err)
	}

	for _, dirEntry := range dirEntries {
		s.Add(1)
		go scanFile(i, dir, dirEntry, s)
	}
}
