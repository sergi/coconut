package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/jinzhu/configor"
	ct "github.com/sergi/coconut/reader"
)

func loadGeoCodes() *os.File {
	codes, err := os.Open("./geocode.csv")
	if err != nil {
		log.Fatal(err)
	}
	return codes
}

func main() {
	config := ct.Config{}
	var dryRun bool

	flag.StringVar(&config.Destination, "destination", "", "Destination Folder")
	flag.BoolVar(&dryRun, "dry-run", false, "Run without copying any files")
	flag.Parse()

	configor.Load(&config, "config.yml")
	// log.SetOutput(os.Stdout)
	flag.Parse()

	start := time.Now()
	l := ct.CreateLocatorFromCSV(loadGeoCodes())
	p := ct.New(flag.Args(), l, &config)

	m := p.Start()

	var paths []string
	for p := range m {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, path := range paths {
		// fmt.Printf("%v %s %s\n", m[path].Time, m[path].Geo, path)

		finalPath, err := ct.ExecuteConfigTemplate(m[path], config.Template)
		if err != nil {
			log.Fatal(err)
		}

		_, file := filepath.Split(path)
		absPath, _ := filepath.Abs(filepath.Join(finalPath, file))
		fmt.Println(absPath)
	}

	color.Green("%d files processed.", p.ProcessedFiles())
	color.Yellow("%d duplicates found.", p.DuplicateFiles())
	fmt.Println(time.Since(start))
}
