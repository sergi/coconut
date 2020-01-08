package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/jinzhu/configor"
	ct "sergimansilla.com/coconut"
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

	flag.StringVar(&config.Destination, "destination", "", "Destination Folder")
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
		fmt.Printf("%v %s %s\n", m[path].Time, m[path].Geo, path)

		finalPath, err := ct.ExecuteConfigTemplate(m[path], config.Template)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("finalPath", finalPath)
	}

	fmt.Printf("%d files processed.\n", p.ProcessedFiles())
	fmt.Printf("%d duplicates found.\n", p.DuplicateFiles())
	fmt.Println(time.Since(start))
}
