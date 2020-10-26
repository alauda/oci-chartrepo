package pkg

import (
	"io/ioutil"

	"github.com/labstack/echo/v4"
)

// GetChartHandler get the charts data api
func GetChartHandler(c echo.Context) error {
	name := c.Param("name")
	data, err := GetChartData(name)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Content-Type", "application/x-tar")
	_, err = c.Response().Write(data)
	return err

}

// GetChartData get the charts data from a chart name
func GetChartData(name string) ([]byte, error) {
	ref := pathToRefCache[name]
	result, err := GlobalBackend.Hub.DownloadBlob(ref.Name, ref.Digest)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(result)
}
