package validation_test

import (
    "io"
    "net/http"
    "testing"
    "strings"
    "errors"
    "time"
    "formlex/bello-validation-go/validation"
)

// ----------------------------------------------------
// 📦 LÓGICA DE MOCKING (Reemplazo de http.RoundTripper)
// ----------------------------------------------------

// MockRoundTripper implementa la interfaz http.RoundTripper.
type MockRoundTripper struct {
    Handler func(*http.Request) (*http.Response, error)
}

// RoundTrip es el método principal que http.Client llamará.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    if m.Handler != nil {
        return m.Handler(req)
    }
    return nil, errors.New("Mock handler not set")
}

// NewMockClient crea un *http.Client que utiliza nuestro MockRoundTripper.
func NewMockClient(handler func(*http.Request) (*http.Response, error)) *http.Client {
    return &http.Client{
        Transport: &MockRoundTripper{Handler: handler},
        Timeout: 10 * time.Second, 
    }
}

// Helper para crear el cuerpo de la respuesta mock
func mockResponseBody(body string, statusCode int) *http.Response {
    return &http.Response{
        StatusCode: statusCode,
        Body:       io.NopCloser(strings.NewReader(body)), // io.NopCloser es la forma moderna
        Header:     make(http.Header),
    }
}

// ----------------------------------------------------
// 🧪 TESTS
// ----------------------------------------------------

func TestValidateDocument_Success(t *testing.T) {
    mockPayload := validation.ValidationPayload{CountryCode: "es", DocumentType: "dni", DocumentNumber: "12345678Z"}
    mockAPIKey := "mock_key_123"
    mockBaseURL := "https://mock.api.com"
    
    successJSON := `{
        "valid": true,
        "message": "Document format is valid."
    }`

    mockHandler := func(req *http.Request) (*http.Response, error) {
        // Validaciones del request
        expectedURL := mockBaseURL + "/v1/validate-document"
        if req.URL.String() != expectedURL {
            t.Errorf("URL incorrecta. Esperada: %s, Obtenida: %s", expectedURL, req.URL.String())
        }
        
        // Devolver la respuesta mock 200 OK
        return mockResponseBody(successJSON, http.StatusOK), nil
    }

    mockClient := NewMockClient(mockHandler)
    sut := validation.NewValidationClient(mockAPIKey, mockBaseURL, mockClient) 

    result, err := sut.ValidateDocument(mockPayload)

    if err != nil {
        t.Fatalf("Esperaba éxito, obtuve error: %v", err)
    }

    if !result.Valid {
        t.Errorf("Resultado incorrecto. Esperaba valid=true, obtuve valid=false")
    }
    if result.Message != "Document format is valid." {
        t.Errorf("Mensaje incorrecto. Obtenido: %s", result.Message)
    }
}

func TestValidateDocument_403UsageLimitExceeded(t *testing.T) {
    mockPayload := validation.ValidationPayload{CountryCode: "es", DocumentType: "dni", DocumentNumber: "12345678Z"}
    mockAPIKey := "mock_key_123"
    mockBaseURL := "https://mock.api.com"
    
    // JSON del error 403 (ajustar a la estructura real de tu API)
    errorJSON := `{"message": "Daily usage limit exceeded."}` 

    mockHandler := func(req *http.Request) (*http.Response, error) {
        return mockResponseBody(errorJSON, http.StatusForbidden), nil
    }

    mockClient := NewMockClient(mockHandler)
    sut := validation.NewValidationClient(mockAPIKey, mockBaseURL, mockClient) 

    _, err := sut.ValidateDocument(mockPayload)

    if err == nil {
        t.Fatalf("Esperaba un error de límite de uso, obtuve nil")
    }
    
    // Asunción: tu SDK envuelve los errores HTTP en un tipo validation.BelloError
    var apiErr *validation.BelloError
    if errors.As(err, &apiErr) {
        if apiErr.StatusCode != http.StatusForbidden {
            t.Errorf("Status Code incorrecto. Esperado: 403, Obtenido: %d", apiErr.StatusCode)
        }
        
        // Aquí se valida el mensaje que tu SDK debería generar a partir del JSON
        expectedMsg := "API returned status 403: Daily usage limit exceeded." 
        if apiErr.Error() != expectedMsg {
            t.Errorf("Mensaje de error incorrecto.\nEsperado: %s\nObtenido: %s", expectedMsg, apiErr.Error())
        }
    } else {
        t.Errorf("El error no es del tipo BelloError. Obtenido: %v", err)
    }
}