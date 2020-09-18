package pkg

import (
	"encoding/json"
	"github.com/heroku/docker-registry-client/registry"
	"io/ioutil"
	"k8s.io/klog"
	"strings"
	"sync"
)

var (

	// backend global var
	GlobalBackend *Backend

	// cache section
	refToChartCache map[string]*HelmOCIConfig

	pathToRefCache map[string]RefData
)

var l = sync.Mutex{}

func init() {
	refToChartCache = make(map[string]*HelmOCIConfig)
	pathToRefCache = make(map[string]RefData)
}

type Backend struct {
	URL string
	Hub *registry.Registry
}

func NewBackend(url string) *Backend {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	hub, err := registry.NewInsecure(url, "", "")
	if err != nil {
		panic(err)
	}

	return &Backend{
		URL: url,
		Hub: hub,
	}
}

// ListObjects parser all helm chart basic info from oci manifest
// skip all manifests that are not helm type
func (b *Backend) ListObjects() ([]HelmOCIConfig, error) {
	repositories, err := b.Hub.Repositories()
	if err != nil {
		return nil, err
	}

	var objects []HelmOCIConfig

	for _, image := range repositories {
		tags, err := b.Hub.Tags(image)
		if err != nil {
			if strings.Contains(err.Error(), "repository name not known to registry") {
				klog.Error("err list tags for repo: ", err)
				continue
			}
			return nil, err
		}
		for _, tag := range tags {
			manifest, err := b.Hub.OCIManifestV1(image, tag)
			if err != nil {
				return nil, err
			}

			// if one tag is not helm, consider this image is not
			if manifest.Config.MediaType != "application/vnd.cncf.helm.config.v1+json" {
				break
			}

			// only one layer is allowed
			if len(manifest.Layers) != 1 {
				break
			}

			ref := image + ":" + tag

			// lookup in cache first
			obj := refToChartCache[ref]
			if obj != nil {
				objects = append(objects, *obj)
				continue
			}

			// fetch manifest config and parse to helm info
			digest := manifest.Config.Digest
			result, err := b.Hub.DownloadBlob(image, digest)
			if err != nil {
				return nil, err
			}
			body, err := ioutil.ReadAll(result)
			if err != nil {
				return nil, err
			}
			result.Close()

			cfg := &HelmOCIConfig{}
			err = json.Unmarshal(body, cfg)
			if err != nil {
				return nil, err
			}

			cfg.Digest = manifest.Layers[0].Digest.Encoded()
			objects = append(objects, *cfg)

			// may be helm and captain are pulling same time
			l.Lock()
			refToChartCache[ref] = cfg
			pathToRefCache[genPath(cfg.Name, cfg.Version)] = RefData{
				Name:   image,
				Digest: manifest.Layers[0].Digest,
			}
			l.Unlock()
		}

	}
	return objects, nil

}
