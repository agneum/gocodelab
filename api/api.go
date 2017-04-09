package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

type API struct {
	echo     *echo.Echo
	bindAddr string
}

func New(bindAddr string) *API {
	a := &API{}
	a.echo = echo.New()
	a.bindAddr = bindAddr

	g := a.echo.Group("/api")
	g.POST("/driver/", a.addDriver)
	g.GET("/driver/:id", a.getDriver)
	g.DELETE("/driver/:id", a.deleteDriver)
	g.GET("/driver/:lat/:lon/nearest", a.nearestDrivers)
	return a
}

func (a *API) Start() error {
	return a.echo.Start(a.bindAddr)
}

func (a *API) addDriver(c echo.Context) error {
	p := &Payload{}
	if err := c.Bind(p); err != nil {
		return c.JSON(http.StatusUnsupportedMediaType, &DefaultResponse{
			Success: false,
			Message: "set content-type application/json or check your payload data",
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

	return c.JSON(http.StatusOK, &DriverResponse{
		Success: true,
		Message: "Found",
		Driver:  id,
	})
}

func (a *API) deleteDriver(c echo.Context) error {
	driverID := c.Param("id")
	_, err := strconv.Atoi(driverID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: "could not convert string to integer",
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

	_, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: "lat. could not parse float",
		})
	}

	_, err = strconv.ParseFloat(lon, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &DefaultResponse{
			Success: false,
			Message: "lon. could not parse float",
		})
	}
	// coordinates processing

	return c.JSON(http.StatusOK, &NearestDriverResponse{
		Success: true,
		Message: "drivers list",
	})
}