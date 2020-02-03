package reader

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"go4.org/media/heif"
)

type MetaData struct {
	CreationDate time.Time
	Location     string
}

func NewMetaData(f *os.File, meta *FilePath, loc *Locator) (MetaData, error) {
	var exifRaw io.Reader
	if meta.ext == ".heic" {
		h := heif.Open(f)
		e, err := h.EXIF()
		if err != nil {
			fmt.Println("EXIF Error: ", meta.path, err)
			return MetaData{}, err
		}
		exifRaw = bytes.NewReader(e)
	} else {
		exifRaw = f
	}

	xif, err := exif.Decode(exifRaw)
	if err != nil {
		fmt.Println("EXIF Error: ", meta.path, err)
		return MetaData{}, err
	}

	time := getCreationDate(f, xif)
	location := getGeo(xif, loc)

	return MetaData{
		time, location,
	}, nil
}

func getCreationDate(f *os.File, x *exif.Exif) time.Time {
	var createdOn time.Time
	createdOn, err := x.DateTime()
	if err != nil {
		file, err := f.Stat()

		if err != nil {
			fmt.Println(err)
		}

		createdOn = file.ModTime()
	}
	return createdOn
}

func getGeo(x *exif.Exif, l *Locator) string {
	var nearestTown string
	lat, long, geoErr := x.LatLong()
	if geoErr != nil {
		nearestTown = "Unknown Location"
	} else {
		nearestTown = l.FindNearest(&LatLng{X: lat, Y: long, ID: 0})
	}
	return nearestTown
}
