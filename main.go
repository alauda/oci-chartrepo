package main

import (
	"flag"
	"net/http"

	"github.com/alauda/oci-chartrepo/pkg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	var registryOpts = &pkg.RegistryOptions{}
	// flags
	port := flag.String("port", "8080", "listen port")
	//TODO: remove in chart args and here
	flag.String("storage", "registry", "storage backend(only registry for now)")
	flag.StringVar(&registryOpts.URL, "storage-registry-repo", "localhost:5000", "oci registry address")
	flag.StringVar(&registryOpts.Scheme, "storage-registry-scheme", "", "oci registry address scheme, default is http")
	flag.Parse()

	// Get registry scheme, and get user info from secret config file(if it exists).
	if err := registryOpts.FullfillRegistryOptions(); err != nil {
		panic(err)
	}

	// Echo instance
	e := echo.New()

	pkg.GlobalBackend = pkg.NewBackend(registryOpts)
	// When multiple instance of oci-chartrepo exist, this will make sure every instance
	// has the internal cache before it gets requrest to individual chart. Of course this will slow down
	// the startup process, we need to add heathcheck later
	// TODO: add health check for pod
	if _, err := pkg.GlobalBackend.ListObjects(); err != nil {
		e.Logger.Fatal("init chart registry cache error", err)
	}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)
	e.GET("/index.yaml", pkg.IndexHandler)
	e.GET("/charts/:name", pkg.GetChartHandler)

	// Start server
	e.Logger.Fatal(e.Start(":" + *port))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
