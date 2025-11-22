package github

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-github/v79/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestClient sets up a test HTTP server and a GitHub client pointing to it.
func setupTestClient(t *testing.T, handler http.Handler) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// The handler passed from the test is registered on the mux
	if handler != nil {
		mux.Handle("/", handler)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create a client and override the base URL to point to the test server
	ghClient, err := github.NewClient(nil).WithEnterpriseURLs(server.URL, server.URL)
	require.NoError(t, err)

	client := &Client{
		client: ghClient,
		logger: logger,
	}

	return client, mux
}

func TestClient_FetchStars(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		client, mux := setupTestClient(t, nil)
		repoName := "owner/repo"
		expectedStars := 1234

		mux.HandleFunc(fmt.Sprintf("/api/v3/repos/%s", repoName), func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"stargazers_count": %d}`, expectedStars)
		})

		repo, err := client.FetchStars(context.Background(), repoName)
		require.NoError(t, err)
		require.NotNil(t, repo)
		assert.Equal(t, repoName, repo.FullName)
		assert.Equal(t, expectedStars, repo.Stars)
	})

	t.Run("Not Found", func(t *testing.T) {
		t.Parallel()
		client, mux := setupTestClient(t, nil)
		repoName := "owner/non-existent-repo"

		mux.HandleFunc(fmt.Sprintf("/api/v3/repos/%s", repoName), func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message": "Not Found"}`)
		})

		repo, err := client.FetchStars(context.Background(), repoName)
		require.Error(t, err)
		assert.Nil(t, repo)
		var fetchErr *FetchError
		assert.ErrorAs(t, err, &fetchErr)
	})

	t.Run("Rate Limited", func(t *testing.T) {
		t.Parallel()
		client, mux := setupTestClient(t, nil)
		repoName := "owner/rate-limited-repo"

		mux.HandleFunc(fmt.Sprintf("/api/v3/repos/%s", repoName), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"message": "API rate limit exceeded"}`)
		})

		repo, err := client.FetchStars(context.Background(), repoName)
		require.Error(t, err)
		assert.Nil(t, repo)
		var rateLimitErr *RateLimitError
		// The go-github library might return a *github.RateLimitError, which our client should wrap.
		// So we check if the underlying cause is a *github.RateLimitError
		var ghRateLimitErr *github.RateLimitError
		require.ErrorAs(t, err, &ghRateLimitErr)
		assert.ErrorAs(t, err, &rateLimitErr, "The error should be wrapped in our custom RateLimitError")
	})

	t.Run("Invalid Repo Name", func(t *testing.T) {
		t.Parallel()
		client, _ := setupTestClient(t, nil)
		repoName := "invalid-repo-name"

		repo, err := client.FetchStars(context.Background(), repoName)
		require.Error(t, err)
		assert.Nil(t, repo)

		// The error should come from our domain validation
		assert.Contains(t, err.Error(), "invalid repository full name format")
	})
}
