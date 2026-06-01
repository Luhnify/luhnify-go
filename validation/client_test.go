package validation_test

import (
    "io"
    "net/http"
    "testing"
    "strings"
    "errors"
    "time"
    "github.com/Luhnify/luhnify-go"
)

// ----------------------------------------------------
// 📦 MOCKING LOGIC (http.RoundTripper Replacement)
// ----------------------------------------------------

// MockRoundTripper implements the http.RoundTripper interface.
type MockRoundTripper struct {
    Handler func(*http.Request) (*http.Response, error)
}

// RoundTrip is the main method that http.Client will call.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    if m.Handler != nil {
        return m.Handler(req)
    }
    return nil, errors.New("Mock handler not set")
}

// NewMockClient creates an *http.Client that uses our MockRoundTripper.
func NewMockClient(handler func(*http.Request) (*http.Response, error)) *http.Client {
    return &http.Client{
        Transport: &MockRoundTripper{Handler: handler},
        Timeout: 10 * time.Second, 
    }
}

// Helper to build a mock response body
func mockResponseBody(body string, statusCode int) *http.Response {
    return &http.Response{
        StatusCode: statusCode,
        Body:       io.NopCloser(strings.NewReader(body)),
        Header:     make(http.Header),
    }
}

// ----------------------------------------------------
// 🧪 TESTS
// ----------------------------------------------------

func TestValidateDocument_Success(t *testing.T) {
    mockPayload := validation.ValidationPayload{CountryCode: "es", DocumentType: "dni", DocumentNumber: "12345678Z"}
    mockAPIKey := "mock_key_123"
    
    successJSON := `{
        "valid": true,
        "message": "Document format is valid."
    }`

    mockHandler := func(req *http.Request) (*http.Response, error) {
        expectedURL := "https://api.luhnify.com/v1/validate"
        if req.URL.String() != expectedURL {
			t.Errorf("incorrect URL target. Expected: %s, Got: %s", expectedURL, req.URL.String())
		}

		// Verify API Key transmission header
		if req.Header.Get("X-API-Key") != mockAPIKey {
			t.Errorf("missing or incorrect X-API-Key header. Got: %s", req.Header.Get("X-API-Key"))
		}
        
        // Devolver la respuesta mock 200 OK
        return mockResponseBody(successJSON, http.StatusOK), nil
    }

    mockClient := NewMockClient(mockHandler)
    sut := validation.NewValidationClient(mockAPIKey, mockClient) 

    result, err := sut.ValidateDocument(mockPayload)

    if err != nil {
		t.Fatalf("expected success response, got unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("incorrect validation outcome. Expected valid=true, got valid=false")
	}
	if result.Message != "Document format is valid." {
		t.Errorf("incorrect response message. Got: %s", result.Message)
	}
}

func TestValidateDocument_403UsageLimitExceeded(t *testing.T) {
	mockPayload := validation.ValidationPayload{CountryCode: "es", DocumentType: "dni", DocumentNumber: "12345678Z"}
	mockAPIKey := "mock_key_123"
    
	errorJSON := `{"message": "Daily usage limit exceeded."}` 

	mockHandler := func(req *http.Request) (*http.Response, error) {
		return mockResponseBody(errorJSON, http.StatusForbidden), nil
	}

	mockClient := NewMockClient(mockHandler)
	sut := validation.NewValidationClient(mockAPIKey, mockClient) 

	_, err := sut.ValidateDocument(mockPayload)

	if err == nil {
		t.Fatalf("expected an architecture usage limit error, got nil")
	}
    
	// Verification targets the new LuhnifyError structure type
	var apiErr *validation.LuhnifyError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode != http.StatusForbidden {
			t.Errorf("incorrect status code mapping. Expected: 403, Got: %d", apiErr.StatusCode)
		}
        
		expectedMsg := "Daily usage limit exceeded." 
		if apiErr.Message != expectedMsg {
			t.Errorf("incorrect structured error message mapping.\nExpected: %s\nGot: %s", expectedMsg, apiErr.Message)
		}
	} else {
		t.Errorf("returned error is not of type LuhnifyError. Got: %v", err)
	}
}