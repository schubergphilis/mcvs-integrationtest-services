package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type User struct {
	Action   string `json:"action"`
	Email    string `json:"email"`
	Facility string `json:"facility"`
	Group    string `json:"group"`
	Name     string `json:"name"`
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Okta!")
	})

	e.POST("/authorization/users", func(c echo.Context) error {
		u := new(User)
		if err := c.Bind(u); err != nil {
			return err
		}

		if u.Facility == u.Group {
			return c.JSON(http.StatusOK, "allowed")
		}

		return c.JSON(http.StatusUnauthorized, "denied")
	})

	e.Logger.Fatal(e.Start(":1323"))
}
