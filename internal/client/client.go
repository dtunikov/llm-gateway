package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DoRequest performs an HTTP request and decodes the JSON response.
func DoRequest(ctx context.Context, client *http.Client, method, url string, headers map[string]string, requestBody interface{}, responseBody interface{}) error {
	var reqBodyBytes []byte
	if requestBody != nil {
		var err error
		reqBodyBytes, err = json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("API returned non-200 status: %d, body: %s", httpResp.StatusCode, respBody)
	}

	if responseBody != nil {
		if err := json.NewDecoder(httpResp.Body).Decode(responseBody); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	return nil
}
