package certification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/redhat-best-practices-for-k8s/checks"
)

const defaultBaseURL = "https://catalog.redhat.com/api/containers/v1"

// PyxisValidator implements checks.CertificationValidator using the Red Hat Pyxis API.
type PyxisValidator struct {
	httpClient *http.Client
	baseURL    string
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
	}
}

// pyxisResponse is the generic Pyxis API response wrapper.
type pyxisResponse struct {
	Data []json.RawMessage `json:"data"`
}

func (v *PyxisValidator) queryPyxis(endpoint string) bool {
	resp, err := v.httpClient.Get(endpoint)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var result pyxisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	return len(result.Data) > 0
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
