package coconut

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"strings"

	"fmt"

	"github.com/kyroy/kdtree"
)

// LatLng ...
type LatLng struct {
	X  float64
	Y  float64
	ID int
}

// Dimensions ...
func (p *LatLng) Dimensions() int {
	return 2
}

// Dimension ...
func (p *LatLng) Dimension(i int) float64 {
	if i == 0 {
		return p.X
	}
	return p.Y
}

// String ...
func (p *LatLng) String() string {
	return fmt.Sprintf("{%.2f %.2f %d}", p.X, p.Y, p.ID)
}

// Locator ...
type Locator struct {
	Locations []string
	tree      *kdtree.KDTree
}

// CreateLocatorFromCSV ...
func CreateLocatorFromCSV(csvReader io.Reader) *Locator {
	r := csv.NewReader(csvReader)
	position := 0
	tree := kdtree.New([]kdtree.Point{})
	nameList := []string{}
	var str strings.Builder
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		lat, _ := strconv.ParseFloat(record[0], 64)
		lng, _ := strconv.ParseFloat(record[1], 64)
		tree.Insert(&LatLng{X: lat, Y: lng, ID: position})
		fmt.Fprintf(&str, "%s-%s", record[3], record[2])
		nameList = append(nameList, str.String())
		position++
		str.Reset()
	}

	return &Locator{nameList, tree}
}

// FindNearest finds nearest town given some coordinates
func (l *Locator) FindNearest(loc *LatLng) string {
	nearestPoint := l.tree.KNN(loc, 1)
	nearestTown := l.Locations[nearestPoint[0].(*LatLng).ID]
	return nearestTown
}
