package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"node-test/internal/master/service"
)

type (
	RouterDependencies struct {
		Logger         zap.Logger
		StorageService service.UploadService
	}
)

//	@title			Talento
//	@version		1.0
//	@description	Talento V1 API

//	@contact.name	Yuri Gasparyan
//	@contact.email	gasparyan.yur@gmail.com

//	@host	localhost:8080

// MakeRoutes create echo routes from dependencies.
//
//nolint:funlen // no another solution
func MakeRoutes(
	dependencies *RouterDependencies,
) *echo.Echo {

	e := echo.New()

	//e.Use(middleware.Secure())
	//e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	//	AllowOrigins:     []string{"https://*", "http://*", "ws://*"},
	//	Skipper:          middleware.DefaultSkipper,
	//	AllowMethods:     middleware.DefaultCORSConfig.AllowMethods,
	//	AllowCredentials: true,
	//}))

	router := e.Group("/api/v1")
	storage := router.Group("/storage")
	{
		storageH := newStorageHandler(dependencies.StorageService)
		storage.Use(middleware.Recover())
		storage.Use(middleware.Logger())
		storage.GET("/ws/upload", storageH.WSUpload)
		storage.GET("/ws/download", storageH.WSDownload)

	}

	return e
}
