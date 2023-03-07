package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	initRouting(e)
	e.Logger.Fatal(e.Start(":1323"))
}

func initRouting(e *echo.Echo) {
	e.GET("/", hello)
	e.GET("/users/:id", getUser)
	e.GET("/show", show)
	e.POST("/save", save)
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World")
}

func getUser(c echo.Context) error {
	id := c.Param("id")
	return c.String(http.StatusOK, id)
}

func show(c echo.Context) error {
	team := c.QueryParam("team")
	member := c.QueryParam("member")
	return c.String(http.StatusOK, "team:" + team + ", member:" + member)
}

func save(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	return c.String(http.StatusOK, "name:" + name + ", email:" + email)
}