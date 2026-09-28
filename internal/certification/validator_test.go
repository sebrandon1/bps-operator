package certification

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestQueryPyxisCachesSuccessfulResponses(t *testing.T) {
	for _, test := range []struct {
		name      string
		data      []json.RawMessage
		certified bool
	}{
		{name: "certified", data: []json.RawMessage{json.RawMessage(`{"id":"123"}`)}, certified: true},
		{name: "not certified", data: []json.RawMessage{}, certified: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var requests atomic.Int32
			server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				_ = json.NewEncoder(w).Encode(pyxisResponse{Data: test.data})
			})
			defer server.Close()

			validator := NewPyxisValidator(server.URL)
			for range 2 {
				if got := validator.queryPyxis(server.URL); got != test.certified {
					t.Fatalf("queryPyxis() = %t, want %t", got, test.certified)
				}
			}
			if got := requests.Load(); got != 1 {
				t.Errorf("got %d API requests, want 1", got)
			}
		})
	}
}

func TestQueryPyxisDoesNotCacheClientErrors(t *testing.T) {
	var requests atomic.Int32
	server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	})
	defer server.Close()

	validator := NewPyxisValidator(server.URL)
	for range 2 {
		if validator.queryPyxis(server.URL) {
			t.Fatal("queryPyxis() = true, want false")
		}
	}
	if got := requests.Load(); got != 2 {
		t.Errorf("got %d API requests, want 2", got)
	}
}

func TestQueryPyxisDoesNotCacheMalformedResponses(t *testing.T) {
	var requests atomic.Int32
	server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte("not-json"))
	})
	defer server.Close()

	validator := NewPyxisValidator(server.URL)
	for range 2 {
		if validator.queryPyxis(server.URL) {
			t.Fatal("queryPyxis() = true, want false")
		}
	}
	if got := requests.Load(); got != 2 {
		t.Errorf("got %d API requests, want 2", got)
	}
}

func TestQueryPyxisCacheExpires(t *testing.T) {
	var requests atomic.Int32
	server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_ = json.NewEncoder(w).Encode(pyxisResponse{Data: []json.RawMessage{json.RawMessage(`{"id":"123"}`)}})
	})
	defer server.Close()

	validator := NewPyxisValidator(server.URL)
	validator.cacheTTL = 10 * time.Millisecond
	first := validator.queryPyxis(server.URL)
	second := validator.queryPyxis(server.URL)
	if !first || !second {
		t.Fatal("expected successful Pyxis responses")
	}
	time.Sleep(20 * time.Millisecond)
	if !validator.queryPyxis(server.URL) {
		t.Fatal("expected successful Pyxis response after cache expiry")
	}
	if got := requests.Load(); got != 2 {
		t.Errorf("got %d API requests, want 2", got)
	}
}

func TestQueryPyxisCoalescesConcurrentRequests(t *testing.T) {
	var requests atomic.Int32
	server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_ = json.NewEncoder(w).Encode(pyxisResponse{Data: []json.RawMessage{json.RawMessage(`{"id":"123"}`)}})
	})
	defer server.Close()

	validator := NewPyxisValidator(server.URL)
	const callers = 20
	var wg sync.WaitGroup
	wg.Add(callers)
	for range callers {
		go func() {
			defer wg.Done()
			if !validator.queryPyxis(server.URL) {
				t.Error("expected successful Pyxis response")
			}
		}()
	}
	wg.Wait()
	if got := requests.Load(); got != 1 {
		t.Errorf("got %d API requests, want 1", got)
	}
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
