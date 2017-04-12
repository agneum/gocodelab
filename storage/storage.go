package storage

import (
	"errors"
	"math"

	"github.com/dhconnelly/rtreego"
)

type (
	Location struct {
		Lat float64
		Lon float64
	}
	Driver struct {
		ID           int
		LastLocation Location
	}
	DriverStorage struct {
		drivers    map[int]*Driver
		localtions *rtreego.Rtree
	}
)

func New() *DriverStorage {
	d := &DriverStorage{}
	d.drivers = make(map[int]*Driver)
	d.localtions = rtreego.NewTree(2, 25, 50)

	return d
}

func (d *DriverStorage) Get(key int) (*Driver, error) {

	driver, ok := d.drivers[key]
	if !ok {
		return nil, errors.New("driver does not exist")
	}

	return driver, nil
}

func (d *DriverStorage) Set(key int, driver *Driver) {
	_, ok := d.drivers[key]
	if !ok {
		d.localtions.Insert(driver)
	}
	d.drivers[key] = driver
}

func (d *DriverStorage) Delete(key int) error {
	driver, ok := d.drivers[key]
	if !ok {
		return errors.New("driver does not exist")
	}

	if d.localtions.Delete(driver) {
		delete(d.drivers, key)
		return nil
	}
	return errors.New("could not remove driver")
}

func (d *DriverStorage) Nearest(number int, lat, lon float64) []*Driver {
	point := rtreego.Point{lat, lon}
	results := d.localtions.NearestNeighbors(number, point)
	var drivers []*Driver

	for _, item := range results {
		if item == nil {
			continue
		}
		drivers = append(drivers, item.(*Driver))
	}

	return drivers
}

func (d *Driver) Bounds() *rtreego.Rect {
	return rtreego.Point{d.LastLocation.Lat, d.LastLocation.Lon}.ToRect(0.01)
}

// Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180
	r = 6378100 // Earth radius in METERS
	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)
	return 2 * r * math.Asin(math.Sqrt(h))
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}
