package docs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScraper_Fetch_GoPackage verifies auto-detection routes "/" packages to FetchGoDocs.
// Given: a library name containing "/"
// When: Fetch is called
// Then: the result comes from the Go docs path
func TestScraper_Fetch_GoPackage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><section id="pkg-overview">Go package docs</section></body></html>`))
	}))
	defer srv.Close()

	scraper := NewScraper(WithGoDocsBaseURL(srv.URL))
	result, err := scraper.Fetch("github.com/spf13/cobra", "api")

	require.NoError(t, err)
	assert.Equal(t, "scraper", result.Source)
	assert.NotEmpty(t, result.Content)
}

// TestScraper_Fetch_NpmPackage verifies auto-detection routes simple names to FetchNpmDocs.
// Given: a library name with no "/" or "."
// When: Fetch is called
// Then: the result comes from npm registry
func TestScraper_Fetch_NpmPackage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"express","description":"web framework","readme":"# Express docs"}`))
	}))
	defer srv.Close()

	scraper := NewScraper(WithNpmRegistryURL(srv.URL))
	result, err := scraper.Fetch("express", "api")

	require.NoError(t, err)
	assert.Equal(t, "scraper", result.Source)
	assert.NotEmpty(t, result.Content)
}

// TestScraper_Fetch_PyPIPackage verifies auto-detection routes "." names to FetchPyPIDocs.
// Given: a library name containing "."
// When: Fetch is called
// Then: the result comes from PyPI
func TestScraper_Fetch_PyPIPackage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"info":{"name":"some.package","summary":"A package","description":"Package docs here."}}`))
	}))
	defer srv.Close()

	scraper := NewScraper(WithPyPIBaseURL(srv.URL))
	result, err := scraper.Fetch("some.package", "api")

	require.NoError(t, err)
	assert.Equal(t, "scraper", result.Source)
}

// TestScraper_FetchGoDocs_HTTPError verifies that HTTP errors from go docs are surfaced.
// Given: a server that returns 404
// When: FetchGoDocs is called
// Then: an error is returned
func TestScraper_FetchGoDocs_HTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	scraper := NewScraper(WithGoDocsBaseURL(srv.URL))
	_, err := scraper.FetchGoDocs("github.com/unknown/pkg")

	require.Error(t, err)
}

// TestScraper_FetchNpmDocs_NoReadme verifies fallback to description when readme is empty.
// Given: npm registry response with empty readme but non-empty description
// When: FetchNpmDocs is called
// Then: the result content equals the description
func TestScraper_FetchNpmDocs_NoReadme(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"mylib","description":"My library description","readme":""}`))
	}))
	defer srv.Close()

	scraper := NewScraper(WithNpmRegistryURL(srv.URL))
	result, err := scraper.FetchNpmDocs("mylib")

	require.NoError(t, err)
	assert.Equal(t, "My library description", result.Content)
}

// TestScraper_FetchPyPIDocs_NoDescription verifies fallback to summary when description is empty.
// Given: PyPI response with empty description but non-empty summary
// When: FetchPyPIDocs is called
// Then: the result content equals the summary
func TestScraper_FetchPyPIDocs_NoDescription(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"info":{"name":"mylib","summary":"Only summary here","description":""}}`))
	}))
	defer srv.Close()

	scraper := NewScraper(WithPyPIBaseURL(srv.URL))
	result, err := scraper.FetchPyPIDocs("mylib")

	require.NoError(t, err)
	assert.Equal(t, "Only summary here", result.Content)
}

// TestScraper_FetchGoDocs_NoOverviewSection verifies fallback to raw body when section is absent.
// Given: HTML with no pkg-overview section
// When: FetchGoDocs is called
// Then: result content is non-empty (raw body fallback)
func TestScraper_FetchGoDocs_NoOverviewSection(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><p>Some other content</p></body></html>`))
	}))
	defer srv.Close()

	scraper := NewScraper(WithGoDocsBaseURL(srv.URL))
	result, err := scraper.FetchGoDocs("github.com/some/pkg")

	require.NoError(t, err)
	assert.NotEmpty(t, result.Content)
}
