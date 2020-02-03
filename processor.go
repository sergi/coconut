package coconut

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/karrick/godirwalk"
	"github.com/patrickmn/go-cache"
)

var allowedExtensions = map[string]bool{".jpg": true, ".jpeg": true, ".cr2": true, ".heic": true}
var amountOfDigesters = 20

type Config struct {
	Template    string `default:"{{.Time | year}}/{{.Time | year}}-{{.Time | month}}/{{.Geo}}"`
	Destination string
}

type FilePath struct {
	path string
	ext  string
}

// Result is the product of processing a file
type Result struct {
	Path     string
	Time     time.Time
	Geo      string
	Checksum []byte
	err      error
}

type Processor struct {
	config  *Config
	folders []string
	DB      *cache.Cache
	l       *Locator
}

func getFileHash(f *os.File) ([]byte, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func New(folders []string, l *Locator, c *Config) Processor {
	return Processor{c, folders, cache.New(0, 0), l}
}

func (p Processor) VisitFolders(folders []string, done <-chan struct{}) (<-chan FilePath, <-chan error) {
	files := make(chan FilePath)
	errc := make(chan error, len(folders))

	var wg sync.WaitGroup
	wg.Add(len(folders))

	walk := func(f string) {
		defer wg.Done()
		// No select needed for this send, since errc is buffered.
		errc <- godirwalk.Walk(f, &godirwalk.Options{
			Callback: p.getVisitorFunc(files, done),
			Unsorted: true,
		})
	}

	for _, folder := range folders {
		go walk(folder)
	}

	go func() {
		wg.Wait()
		close(files)
	}()

	return files, errc
}

func (p Processor) getVisitorFunc(files chan FilePath, done <-chan struct{}) func(string, *godirwalk.Dirent) error {
	return func(path string, de *godirwalk.Dirent) error {
		ext := strings.ToLower(filepath.Ext(path))
		if !de.IsRegular() {
			return nil
		}
		if _, ok := FileExtensions[ext]; !ok {
			return nil
		}
		p.DB.IncrementInt("processedFiles", 1)

		select {
		case files <- FilePath{path, ext}:
		case <-done:
			return errors.New("walk canceled")
		}
		return nil
	}
}

func closeFile(f *os.File) {
	err := f.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// processor reads path names from paths and sends digests of the corresponding
// files on c until either paths or done is closed.
func (p Processor) process(done <-chan struct{}, files <-chan FilePath, c chan<- Result, loc *Locator) {
	processFile := func(m FilePath) {
		path := m.path
		f, err := os.Open(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer closeFile(f)

		h, err := getFileHash(f)
		if err != nil {
			fmt.Println(err)
			return
		}

		checksum := fmt.Sprintf("%x", h)
		duplicate, alreadyProcessed := p.DB.Get(checksum)
		if alreadyProcessed {
			p.DB.IncrementInt("duplicateFiles", 1)
			color.Yellow("Duplicated file\n - %s\n - %s", path, duplicate)
			return
		}

		p.DB.Set(checksum, path, cache.DefaultExpiration)

		f.Seek(0, 0)

		meta, err := NewMetaData(f, &m, loc)
		if err != nil {
			fmt.Println(err)
			return
		}

		select {
		case c <- Result{m.path, meta.CreationDate, meta.Location, h, err}:
		case <-done:
			return
		}
	}

	for f := range files {
		processFile(f)
	}
}

// ProcessAll reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, ProcessAll returns an error.  In that case,
// ProcessAll does not wait for inflight read operations to complete.
func (p Processor) ProcessAll(folders []string, loc *Locator) (map[string]Result, error) {
	// ProcessAll closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.
	done := make(chan struct{})
	defer close(done)

	paths, errc := p.VisitFolders(folders, done)

	results := make(chan Result)
	var wg sync.WaitGroup
	wg.Add(amountOfDigesters)
	for i := 0; i < amountOfDigesters; i++ {
		go func() {
			p.process(done, paths, results, loc)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	// End of pipeline. OMIT

	m := make(map[string]Result)
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}

		m[result.Path] = result
	}

	if err := <-errc; err != nil {
		return nil, err
	}

	return m, nil
}

func (p Processor) Reset() {
	p.DB.Flush()
}
func (p Processor) ProcessedFiles() int {
	v, found := p.DB.Get("processedFiles")
	if !found {
		return 0
	}
	return v.(int)
}
func (p Processor) DuplicateFiles() int {
	v, found := p.DB.Get("duplicateFiles")
	if !found {
		return 0
	}
	return v.(int)
}

// Start here
func (p Processor) Start() map[string]Result {
	p.Reset()
	p.DB.Set("processedFiles", 0, 0)
	p.DB.Set("duplicateFiles", 0, 0)
	m, err := p.ProcessAll(p.folders, p.l)
	if err != nil {
		log.Fatal(err)
	}

	return m
}
