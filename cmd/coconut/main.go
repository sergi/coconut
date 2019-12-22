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
	// flag.StringVar(&Config.DB.Name, "db-name", "", "database name")
	// flag.StringVar(&Config.DB.User, "db-user", "root", "database user")
	flag.Parse()

	configor.Load(&config, "config.yml")
	// log.SetOutput(os.Stdout)
	flag.Parse()

	start := time.Now()
	l := ct.CreateLocatorFromCSV(loadGeoCodes())
	// path := os.Args[1]
	p := ct.New(flag.Args(), l, &config)
	// fmt.Println(time.Since(start))

	m := p.Start()

	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%v %s %s\n", m[path].Time, m[path].Geo, path)

		finalPath, err := ct.ExecuteConfigTemplate(m[path], config.Template)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(finalPath)
	}

	fmt.Printf("%d files processed.\n", p.ProcessedFiles())
	fmt.Printf("%d duplicates found.\n", p.DuplicateFiles())
	fmt.Println(time.Since(start))
}
