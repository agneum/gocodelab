package api

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/agneum/gocodelab/storage"
	"github.com/dhconnelly/rtreego"
	"github.com/labstack/echo"
)

type API struct {
	database  *storage.DriverStorage
	waitGroup sync.WaitGroup
	echo      *echo.Echo
	bindAddr  string
}

func New(bindAddr string, lruSize int) *API {
	a := &API{}
	a.database = storage.New(lruSize)
	a.echo = echo.New()
	a.bindAddr = bindAddr

	g := a.echo.Group("/api")
	g.POST("/driver/", a.addDriver)
	g.GET("/driver/:id", a.getDriver)
	g.DELETE("/driver/:id", a.deleteDriver)
	g.GET("/driver/:lat/:lon/nearest", a.nearestDrivers)

	return a
}

func (a *API) Start() {
	a.waitGroup.Add(1)
	go func() {
		a.echo.Start(a.bindAddr)
		a.waitGroup.Done()
	}()
	a.waitGroup.Add(1)
	go a.removeExpired()
}

func (a *API) WaitStop() {
	a.waitGroup.Wait()
}

func (a *API) removeExpired() {
	for range time.Tick(1) {
		a.database.DeleteExpired()
	}
}

func (a *API) addDriver(c echo.Context) error {
	p := &Payload{}
	if err := c.Bind(p); err != nil {
		return c.JSON(http.StatusUnsupportedMediaType, &DefaultResponse{
			Success: false,
			Message: "set content-type application/json or check your payload data",
		})
	}
	driver := &storage.Driver{}
	driver.ID = p.DriverID
	driver.LastLocation = storage.Location{
		Lat: p.Location.Latitude,
		Lon: p.Location.Longitude,
	}

	if err := a.database.Set(driver); err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, &DefaultResponse{
		Success: true,
		Message: "the driver has been added",
	})
}

func (a *API) getDriver(c echo.Context) error {
	driverID := c.Param("id")
	id, err := strconv.Atoi(driverID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: "driver not found. could not convert string to integer",
		})
	}

	d, err := a.database.Get(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, &DriverResponse{
		Success: true,
		Message: "Found",
		Driver:  d,
	})
}

func (a *API) deleteDriver(c echo.Context) error {
	driverID := c.Param("id")
	key, err := strconv.Atoi(driverID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: "could not convert string to integer",
		})
	}
	err = a.database.Delete(key)

	if err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, &DefaultResponse{
		Success: true,
		Message: "driver has been removed",
	})
}

func (a *API) nearestDrivers(c echo.Context) error {
	lat := c.Param("lat")
	lon := c.Param("lon")
	if lat == "" || lon == "" {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: "coordinates are empty",
		})
	}

	latValue, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: "lat. could not parse float",
		})
	}

	lonValue, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: "lon. could not parse float",
		})
	}

	drivers := a.database.Nearest(rtreego.Point{latValue, lonValue}, 10)

	return c.JSON(http.StatusOK, &NearestDriverResponse{
		Success: true,
		Message: "drivers list",
		Drivers: drivers,
	})
}
