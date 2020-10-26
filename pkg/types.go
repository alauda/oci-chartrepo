package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/klog"
)

const (
	// SecretCfgPath should in JSON format, is the kubernetes.io/dockerconfigjson types of kubernetes secret.
	// The content includes your private docker registry FQDN, username, password, email.
	SecretCfgPath = "/etc/secret/dockerconfigjson"

	// SchemeTypeHTTP defines const "http" for registry URL scheme
	SchemeTypeHTTP = "http"
	// SchemeTypeHTTPS defines const "https" for registry URL scheme
	SchemeTypeHTTPS = "https"

	// PrefixHTTP defines const SchemeTypeHTTP+"://"
	PrefixHTTP = SchemeTypeHTTP + "://"
	// PrefixHTTPS defines const SchemeTypeHTTPS+"://"
	PrefixHTTPS = SchemeTypeHTTPS + "://"
)

// RegistryOptions defines the options for oci registry
type RegistryOptions struct {
	Scheme   string // http or https
	URL      string
	Username string
	Password string
}

// FullfillRegistryOptions get user info from SecretCfgPath.
// The o.scheme is used to connect to the registry.
func (o *RegistryOptions) FullfillRegistryOptions() error {
	body, err := ioutil.ReadFile(SecretCfgPath)
	if err != nil {
		klog.Warningf("Try to read secret file : %s but failed, the reason is : %s", SecretCfgPath, err.Error())
	} else {
		var cfg DockerSecretCfg
		if err := json.Unmarshal(body, &cfg); err != nil {
			return err
		}

		for k, v := range cfg.Auths {
			if o.matchURL(k) {
				o.Username = v.Username
				o.Password = v.Password

				break
			}
		}
	}

	return nil
}

// IsSchemeValid returns o.Scheme is http or https, or not
func (o *RegistryOptions) IsSchemeValid() bool {
	return o.Scheme == SchemeTypeHTTP || o.Scheme == SchemeTypeHTTPS
}

// ValidateAndSetScheme validate scheme from o.Scheme first.
// if o.Scheme is empty or other value, then get scheme from o.URL.
// if none of the above, infer the scheme.
func (o *RegistryOptions) ValidateAndSetScheme() {
	o.Scheme = strings.ToLower(o.Scheme)
	if o.IsSchemeValid() {
		return
	}

	if strings.HasPrefix(o.URL, PrefixHTTPS) {
		o.Scheme = SchemeTypeHTTPS
	} else if strings.HasPrefix(o.URL, PrefixHTTP) {
		o.Scheme = SchemeTypeHTTP
	} else {
		// Do nothing, need to try to infer
	}
}

func (o *RegistryOptions) matchURL(target string) bool {
	source := o.URL
	if strings.HasPrefix(source, PrefixHTTP) {
		source = source[len(PrefixHTTP):]
	} else if strings.HasPrefix(source, PrefixHTTPS) {
		source = source[len(PrefixHTTPS):]
	}

	if strings.HasPrefix(target, PrefixHTTP) {
		target = target[len(PrefixHTTP):]
	} else if strings.HasPrefix(target, PrefixHTTPS) {
		target = target[len(PrefixHTTPS):]
	}

	return source == target
}

// TryToNewRegistry first try to connect to the registry using http.
// If http fails, try to connect with https.
func (o *RegistryOptions) TryToNewRegistry() (*registry.Registry, error) {
	tryURL := fmt.Sprintf("http://%s", o.URL)
	klog.Infof("Try to connect to the registry using HTTP scheme : %s.\n", tryURL)
	r, err := registry.NewInsecure(tryURL, o.Username, o.Password)
	if err != nil {
		klog.Warning("Failed to connect to the registry using HTTP scheme.", err)

		tryURL = fmt.Sprintf("https://%s", o.URL)
		klog.Infof("Try to connect to the registry using HTTPS scheme : %s.\n", tryURL)
		r, err = registry.NewInsecure(tryURL, o.Username, o.Password)
		if err != nil {
			panic(err)
		} else {
			klog.Infof("Successfully connected to the registry : %s.", tryURL)
			o.URL = tryURL
		}
	} else {
		klog.Infof("Successfully connected to the registry : %s.", tryURL)
		o.URL = tryURL
	}

	return r, nil
}

// DockerSecretCfg defines the structure of dockerconfigjson
type DockerSecretCfg struct {
	Auths map[string]SecretCfg `json:"auths"`
}

// SecretCfg defines the structure that in DockerSecretCfg struct
type SecretCfg struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Auth     string `json:"auth"`
}

// RefData defines the structure that contains the name and digist
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
	AppVersion  string `json:"appVersion"`
	Type        string `json:"type"`

	// use first layer of content's now.
	//TODO: make sure this is ok
	Digest string `json:"-"`
}

// ToChartVersion convert HelmOCIConfig to repo.ChartVersion
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
