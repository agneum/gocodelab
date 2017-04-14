package storage

import (
	// "errors"
	"math"
	"sync"
	"time"

	"github.com/agneum/gocodelab/storage/lru"
	"github.com/dhconnelly/rtreego"
	"github.com/pkg/errors"
)

type (
	Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
	Driver struct {
		ID           int      `json:"id"`
		LastLocation Location `json:"location"`
		Expiration   int64    `json:"-"`
		Locations    *lru.LRU `json:"-"`
	}
	DriverStorage struct {
		mu         *sync.RWMutex
		drivers    map[int]*Driver
		localtions *rtreego.Rtree
		lruSize    int
	}
)

func New(lruSize int) *DriverStorage {
	s := new(DriverStorage)
	s.drivers = make(map[int]*Driver)
	s.localtions = rtreego.NewTree(2, 25, 50)
	s.mu = new(sync.RWMutex)
	s.lruSize = lruSize

	return s
}

func (s *DriverStorage) Get(key int) (*Driver, error) {
	s.mu.RLock()
	s.mu.RUnlock()

	driver, ok := s.drivers[key]
	if !ok {
		return nil, errors.New("driver does not exist")
	}

	return driver, nil
}

func (s *DriverStorage) Set(driver *Driver) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.drivers[driver.ID]
	if !ok {
		d = driver
		cache, err := lru.New(s.lruSize)
		if err != nil {
			return errors.Wrap(err, "could not create LRU")
		}
		d.Locations = cache
		s.localtions.Insert(d)
	}
	d.LastLocation = driver.LastLocation
	d.Locations.Add(time.Now().UnixNano(), d.LastLocation)
	d.Expiration = driver.Expiration

	s.drivers[driver.ID] = driver
	return nil
}

func (s *DriverStorage) Delete(key int) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	driver, ok := s.drivers[key]
	if !ok {
		return errors.New("driver does not exist")
	}

	if s.localtions.Delete(driver) {
		delete(s.drivers, key)
		return nil
	}
	return errors.New("could not remove driver")
}

func (s *DriverStorage) Nearest(point rtreego.Point, number int) []*Driver {
	s.mu.Lock()
	defer s.mu.Unlock()

	// point := rtreego.Point{lat, lon}
	results := s.localtions.NearestNeighbors(number, point)
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

func (d *Driver) Expired() bool {
	if d.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > d.Expiration
}

func (s *DriverStorage) DeleteExpired() {
	now := time.Now().UnixNano()
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range s.drivers {
		if v.Expiration > 0 && now > v.Expiration {
			deleted := s.localtions.Delete(v)
			if deleted {
				delete(s.drivers, v.ID)
			}
		}
	}
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
