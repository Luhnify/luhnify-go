// 📄 validation/client.go
package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ValidationClient struct {
	apiKey    string
	httpClient HTTPClient
}

// NewValidationClient es el constructor
func NewValidationClient(apiKey string, client HTTPClient) *ValidationClient {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &ValidationClient{
		apiKey:    apiKey,
		httpClient: client,
	}
}


func (c *ValidationClient) ValidateDocument(payload ValidationPayload) (*ValidationResult, error) {
    
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("failed to serialize payload: %w", err)
    }

	req, err := http.NewRequest(http.MethodPost, "https://api.luhnify.com/v1/validate", bytes.NewBuffer(payloadBytes))
    if err != nil {
        return nil, fmt.Errorf("failed to create http request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-KEY", c.apiKey)
    
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}
	defer resp.Body.Close()

    // --- HTTP Error Handling Logic ---
    if resp.StatusCode != http.StatusOK {
        errorBody, _ := io.ReadAll(resp.Body) 
        
        // Try to decode API standard error response structure
        var apiMsg APIMessage
        if err := json.Unmarshal(errorBody, &apiMsg); err == nil && apiMsg.Message != "" {
            return nil, &BelloError{
                StatusCode: resp.StatusCode,
                Message:    apiMsg.Message,
            }
        }
        
        // Fallback to raw body text if JSON unmarshalling fails
        return nil, &BelloError{
            StatusCode: resp.StatusCode,
            Message:    string(errorBody), 
        }
    }

    // --- Success Handling Logic ---
    var result ValidationResult
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode validation result response: %w", err)
    }
    
    return &result, nil
}