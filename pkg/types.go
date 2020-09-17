package pkg

import (
	"github.com/opencontainers/go-digest"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
)

type RefData struct {
	Name string
	//  digest of the data layer
	Digest digest.Digest

}



// HelmOCIConfig ... from oci manifest config
type HelmOCIConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	APIVersion  string `json:"apiVersion"`
	AppVersion  string `json:"appVersion"'`
	Type        string `json:"type"`

	// use first layer of content's now.
	//TODO: make sure this is ok
	Digest string `json:"-"`
}

func (h *HelmOCIConfig) ToChartVersion() *repo.ChartVersion {

	m := chart.Metadata{}
	m.Version = h.Version
	m.Name = h.Name
	m.APIVersion = h.APIVersion
	m.AppVersion = h.AppVersion
	m.Description = h.Description

	v := repo.ChartVersion{Metadata: &m}
	v.Digest = h.Digest
	v.URLs = []string{"charts/" + genPath(h.Name, h.Version)}

	return &v
}
