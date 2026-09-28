package certification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/redhat-best-practices-for-k8s/checks"
	"golang.org/x/sync/singleflight"
)

const (
	retryAttempts         = 3
	retryBaseDelay        = 500 * time.Millisecond
	certificationCacheTTL = 5 * time.Minute
)

const defaultBaseURL = "https://catalog.redhat.com/api/containers/v1"

// PyxisValidator implements checks.CertificationValidator using the Red Hat Pyxis API.
type PyxisValidator struct {
	httpClient *http.Client
	baseURL    string
	cacheMu    sync.Mutex
	cache      map[string]cacheEntry
	inflight   singleflight.Group
	cacheTTL   time.Duration
}

type cacheEntry struct {
	certified bool
	expiresAt time.Time
}

// Ensure PyxisValidator implements the interface.
var _ checks.CertificationValidator = (*PyxisValidator)(nil)

// NewPyxisValidator creates a PyxisValidator with the given base URL.
// If baseURL is empty, the default Red Hat Catalog API URL is used.
func NewPyxisValidator(baseURL string) *PyxisValidator {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &PyxisValidator{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		cache:      make(map[string]cacheEntry),
		cacheTTL:   certificationCacheTTL,
	}
}

// pyxisResponse is the generic Pyxis API response wrapper.
type pyxisResponse struct {
	Data []json.RawMessage `json:"data"`
}

func (v *PyxisValidator) queryPyxis(endpoint string) bool {
	if certified, ok := v.cachedPyxis(endpoint); ok {
		return certified
	}
	result, _, _ := v.inflight.Do(endpoint, func() (any, error) {
		if certified, ok := v.cachedPyxis(endpoint); ok {
			return certified, nil
		}

		certified, cacheable := v.fetchPyxis(endpoint)
		if cacheable {
			v.cachePyxis(endpoint, certified)
		}
		return certified, nil
	})
	return result.(bool)
}

func (v *PyxisValidator) cachedPyxis(endpoint string) (bool, bool) {
	v.cacheMu.Lock()
	defer v.cacheMu.Unlock()

	entry, ok := v.cache[endpoint]
	if !ok {
		return false, false
	}
	if !time.Now().Before(entry.expiresAt) {
		delete(v.cache, endpoint)
		return false, false
	}
	return entry.certified, true
}

func (v *PyxisValidator) cachePyxis(endpoint string, certified bool) {
	now := time.Now()
	v.cacheMu.Lock()
	defer v.cacheMu.Unlock()

	for cachedEndpoint, entry := range v.cache {
		if !now.Before(entry.expiresAt) {
			delete(v.cache, cachedEndpoint)
		}
	}
	v.cache[endpoint] = cacheEntry{
		certified: certified,
		expiresAt: now.Add(v.cacheTTL),
	}
}

// fetchPyxis performs the request and reports whether the response is safe to cache.
func (v *PyxisValidator) fetchPyxis(endpoint string) (bool, bool) {
	delay := retryBaseDelay
	for attempt := range retryAttempts {
		resp, err := v.httpClient.Get(endpoint)
		if err != nil {
			if attempt < retryAttempts-1 {
				time.Sleep(delay)
				delay *= 2
			}
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var result pyxisResponse
			decodeErr := json.NewDecoder(resp.Body).Decode(&result)
			_ = resp.Body.Close()
			if decodeErr != nil {
				return false, false
			}
			return len(result.Data) > 0, true
		}

		_ = resp.Body.Close()
		// Retry on server errors and rate limiting; fail fast on client errors.
		if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return false, false
		}
		if attempt < retryAttempts-1 {
			time.Sleep(delay)
			delay *= 2
		}
	}
	return false, false
}

// IsContainerCertified checks if a container image is certified by digest.
func (v *PyxisValidator) IsContainerCertified(registry, repository, tag, digest string) bool {
	if digest == "" {
		return false
	}
	endpoint := fmt.Sprintf("%s/repositories/registry/%s/repository/%s/images?filter=docker_image_digest==%s",
		v.baseURL,
		url.PathEscape(registry),
		url.PathEscape(repository),
		url.QueryEscape(digest),
	)
	return v.queryPyxis(endpoint)
}

// IsOperatorCertified checks if an operator is certified for the given OCP version.
func (v *PyxisValidator) IsOperatorCertified(csvName, ocpVersion string) bool {
	if csvName == "" {
		return false
	}
	filter := fmt.Sprintf("csv_name==%s", url.QueryEscape(csvName))
	if ocpVersion != "" {
		filter += fmt.Sprintf(";ocp_version==%s", url.QueryEscape(ocpVersion))
	}
	endpoint := fmt.Sprintf("%s/operators/bundles?filter=%s", v.baseURL, filter)
	return v.queryPyxis(endpoint)
}

// IsHelmChartCertified checks if a Helm chart is certified.
func (v *PyxisValidator) IsHelmChartCertified(chartName, chartVersion, kubeVersion string) bool {
	if chartName == "" {
		return false
	}
	filter := fmt.Sprintf("chart_name==%s", url.QueryEscape(chartName))
	if chartVersion != "" {
		filter += fmt.Sprintf(";version==%s", url.QueryEscape(chartVersion))
	}
	endpoint := fmt.Sprintf("%s/charts?filter=%s", v.baseURL, filter)
	return v.queryPyxis(endpoint)
}
