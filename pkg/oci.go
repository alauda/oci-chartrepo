package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/heroku/docker-registry-client/registry"
	"k8s.io/klog"
)

var (

	// GlobalBackend backend global var
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

// Backend defines a RUL address and a Registry client
type Backend struct {
	URL string
	Hub *registry.Registry
}

// NewBackend create a Registry client and return a Backend structure
func NewBackend(opts *RegistryOptions) *Backend {
	var hub *registry.Registry
	var err error
	opts.ValidateAndSetScheme()
	klog.Infof("registry scheme is %s", opts.Scheme)

	if !opts.IsSchemeValid() {
		hub, err = opts.TryToNewRegistry()
		if err != nil {
			panic(err)
		}
	} else {
		prefix := opts.Scheme + "://"
		if !strings.HasPrefix(opts.URL, prefix) {
			opts.URL = fmt.Sprintf("%s%s", prefix, opts.URL)
		}

		hub, err = registry.NewInsecure(opts.URL, opts.Username, opts.Password)
		if err != nil {
			panic(err)
		}
	}

	return &Backend{
		URL: opts.URL,
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
			klog.Error("err list tags for repo: ", err)
			// You can list Repositories, but the API returns UNAUTHORIZED or PROJECT_POLICY_VIOLATION when you list tags for a repository
			if strings.Contains(err.Error(), "repository name not known to registry") ||
				strings.Contains(err.Error(), "UNAUTHORIZED") ||
				strings.Contains(err.Error(), "PROJECT_POLICY_VIOLATION") {

				continue
			}
			return nil, err
		}
		for _, tag := range tags {
			manifest, err := b.Hub.OCIManifestV1(image, tag)
			if err != nil {
				klog.Warning("err get manifest for tag: ", err)
				// You can list tags, but the API returns UNAUTHORIZED or PROJECT_POLICY_VIOLATION when you get manifest for a tag
				if strings.Contains(err.Error(), "UNAUTHORIZED") ||
					strings.Contains(err.Error(), "PROJECT_POLICY_VIOLATION") {
					break
				}

				// FIXME: continue or return error.
				continue
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
