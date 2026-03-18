package certification

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestIsContainerCertified_Found(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := pyxisResponse{Data: []json.RawMessage{[]byte(`{"id":"123"}`)}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	v := NewPyxisValidator(server.URL)
	if !v.IsContainerCertified("registry.example.com", "repo/image", "latest", "sha256:abc123") {
		t.Error("expected container to be certified")
	}
}

func TestIsContainerCertified_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := pyxisResponse{Data: []json.RawMessage{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	v := NewPyxisValidator(server.URL)
	if v.IsContainerCertified("registry.example.com", "repo/image", "latest", "sha256:abc123") {
		t.Error("expected container to not be certified")
	}
}

func TestIsContainerCertified_EmptyDigest(t *testing.T) {
	v := NewPyxisValidator("http://unused")
	if v.IsContainerCertified("registry", "repo", "tag", "") {
		t.Error("expected false for empty digest")
	}
}

func TestIsOperatorCertified_Found(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := pyxisResponse{Data: []json.RawMessage{[]byte(`{"id":"456"}`)}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	v := NewPyxisValidator(server.URL)
	if !v.IsOperatorCertified("my-operator.v1.0.0", "4.14") {
		t.Error("expected operator to be certified")
	}
}

func TestIsOperatorCertified_Empty(t *testing.T) {
	v := NewPyxisValidator("http://unused")
	if v.IsOperatorCertified("", "4.14") {
		t.Error("expected false for empty CSV name")
	}
}

func TestIsHelmChartCertified_Found(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := pyxisResponse{Data: []json.RawMessage{[]byte(`{"id":"789"}`)}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	v := NewPyxisValidator(server.URL)
	if !v.IsHelmChartCertified("my-chart", "1.0.0", "1.28") {
		t.Error("expected helm chart to be certified")
	}
}

func TestIsHelmChartCertified_Empty(t *testing.T) {
	v := NewPyxisValidator("http://unused")
	if v.IsHelmChartCertified("", "1.0.0", "1.28") {
		t.Error("expected false for empty chart name")
	}
}

func TestIsHelmChartCertified_APIError(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	v := NewPyxisValidator(server.URL)
	if v.IsHelmChartCertified("my-chart", "1.0.0", "1.28") {
		t.Error("expected false on API error")
	}
}
