package storage

import (
	"errors"
	"testing"
	"time"

	"github.com/dhconnelly/rtreego"
	"github.com/stretchr/testify/assert"
)

func TestDriverStorage(t *testing.T) {
	s := New(10)
	s.Set(&Driver{
		ID: 1,
		LastLocation: Location{
			Lat: 1,
			Lon: 1,
		},
		Expiration: time.Now().Add(15).UnixNano(),
	})
	// driver.ID, driver)
	d, err := s.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, d.ID, 1)
	err = s.Delete(1)
	assert.NoError(t, err)
	d, err = s.Get(1)
	assert.Equal(t, err, errors.New("driver does not exist"))
}

func TestNearest(t *testing.T) {
	s := New(10)
	s.Set(&Driver{
		ID: 123,
		LastLocation: Location{
			Lat: 1,
			Lon: 1,
		},
		Expiration: time.Now().Add(15).UnixNano(),
	})
	s.Set(&Driver{
		ID: 321,
		LastLocation: Location{
			Lat: 42.875799,
			Lon: 74.588279,
		},
		Expiration: time.Now().Add(15).UnixNano(),
	})
	s.Set(&Driver{
		ID: 666,
		LastLocation: Location{
			Lat: 42.875799,
			Lon: 74.588279,
		},
		Expiration: time.Now().Add(15).UnixNano(),
	})
	s.Set(&Driver{
		ID: 2319,
		LastLocation: Location{
			Lat: 42.874942,
			Lon: 74.585908,
		},
		Expiration: time.Now().Add(15).UnixNano(),
	})
	s.Set(&Driver{
		ID: 991,
		LastLocation: Location{
			Lat: 42.875744,
			Lon: 74.584503,
		},
		Expiration: time.Now().Add(15).UnixNano(),
	})
	drivers := s.Nearest(rtreego.Point{42.876420, 74.588332}, 3)
	assert.Equal(t, len(drivers), 3)
	assert.Equal(t, drivers[0].ID, 123)
	assert.Equal(t, drivers[1].ID, 321)
	assert.Equal(t, drivers[2].ID, 666)
}

func TestExpire(t *testing.T) {
	s := New(10)
	driver := &Driver{
		ID: 123,
		LastLocation: Location{
			Lat: 42.875744,
			Lon: 74.584503,
		},
		Expiration: time.Now().Add(-15).UnixNano(),
	}
	s.Set(driver)
	s.DeleteExpired()
	d, err := s.Get(123)
	assert.Error(t, err)
	assert.NotEqual(t, d, driver)
}

func BenchmarkNearest(b *testing.B) {
	s := New(100)
	for i := 0; i < 10; i++ {
		s.Set(&Driver{
			ID: i,
			LastLocation: Location{
				Lon: float64(i),
				Lat: float64(i),
			},
			Expiration: time.Now().Add(15).UnixNano(),
		})
	}
	for i := 0; i < b.N; i++ {
		s.Nearest(rtreego.Point{11, 11}, 10)
	}
}
