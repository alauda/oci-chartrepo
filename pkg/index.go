package pkg

import (
	"github.com/ghodss/yaml"
	"github.com/labstack/echo/v4"
	"helm.sh/helm/v3/pkg/repo"
)

// Handler
func IndexHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/x-yaml")
	data, err := genIndex()
	if err != nil {
		return err
	}
	_, err = c.Response().Write(data)
	return err
}

func genIndex() ([]byte, error) {
	index := repo.NewIndexFile()

	objects, err := GlobalBackend.ListObjects()
	if err != nil {
		return nil, err
	}
	index.Merge(createIndex(objects))

	return yaml.Marshal(index)
}

// createIndex create a repo index from list of objects
// 1. merge same name chart
// 2. create index
func createIndex(objects []HelmOCIConfig) *repo.IndexFile {
	index := repo.NewIndexFile()
	for _, object := range objects {
		versions, ok := index.Entries[object.Name]
		if ok {
			versions = append(versions, object.ToChartVersion())
			index.Entries[object.Name] = versions
		} else {
			index.Entries[object.Name] = repo.ChartVersions{object.ToChartVersion()}
		}
	}
	return index
}
