package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoRequestSuccess(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-header-value", r.Header.Get("X-Test-Header"))

		var reqBody map[string]string
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Equal(t, "test_value", reqBody["test_field"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"response_field": "response_value"})
	}))
	defer server.Close()

	// Test data
	ctx := context.Background()
	client := server.Client()
	url := server.URL
	headers := map[string]string{
		"Content-Type":  "application/json",
		"X-Test-Header": "test-header-value",
	}
	requestBody := map[string]string{"test_field": "test_value"}
	var responseBody map[string]string

	// Perform the request
	err := DoRequest(ctx, client, "POST", url, headers, requestBody, &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "response_value", responseBody["response_field"])
}

func TestDoRequestErrorStatus(t *testing.T) {
	// Mock server returning an error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	ctx := context.Background()
	client := server.Client()
	url := server.URL
	headers := map[string]string{}
	var responseBody interface{}

	err := DoRequest(ctx, client, "GET", url, headers, nil, &responseBody)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API returned non-200 status: 500, body: Internal Server Error")
}

func TestDoRequestInvalidJSON(t *testing.T) {
	// Mock server returning invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "invalid json")
	}))
	defer server.Close()

	ctx := context.Background()
	client := server.Client()
	url := server.URL
	headers := map[string]string{}
	var responseBody map[string]string

	err := DoRequest(ctx, client, "GET", url, headers, nil, &responseBody)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response body")
}
