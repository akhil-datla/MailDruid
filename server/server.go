package server

import (
	"fmt"
	"main/confidential"
	"net/http"

	"github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

var e *echo.Echo

func Start(port string, log bool) {
	e = echo.New()
	e.HideBanner = true

	if log {
		e.Use(middleware.Logger())
	}
	e.Use(middleware.Recover())

	DefaultCORSConfig := middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
	}

	e.Use(middleware.CORSWithConfig(DefaultCORSConfig))

	InitializeRoutes()

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

func InitializeRoutes() {
	e.POST("/user/create", createUser)
	e.POST("/user/login", loginUser)

	e.GET("/tasklist", getTaskList)

	r := e.Group("/restricted")
	config := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: confidential.SigningKey,
	}
	r.Use(middleware.JWTWithConfig(config))

	r.GET("/user/info", getUser)
	r.POST("/user/update", updateUser)
	r.DELETE("/user/delete", deleteUser)

	r.GET("/user/folders", getFolders)
	r.POST("/user/update/folder", updateFolder)

	r.POST("/user/update/tags", updateTags)

	r.POST("/user/update/summarycount", updateSummaryCount)

	r.POST("/user/schedule/new", newTask)
	r.POST("/user/schedule/update", updateTask)
	r.DELETE("/user/schedule/delete", deleteTask)

	r.GET("/user/generate", generateSummaryandWordCloud)

}
