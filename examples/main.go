package main

import (
    "fmt"
    "formlex/bello-validation-go/validation"
)

func main() {
    apiKey := "TU_API_KEY_AQUI"
    baseURL := "[https://api.bello.com](https://api.bello.com)"

    // Crear un cliente (usa http.DefaultClient si no se proporciona otro)
    client := validation.NewValidationClient(apiKey, baseURL, nil)

    payload := validation.ValidationPayload{
        CountryCode:    "es",
        DocumentType:   "dni",
        DocumentNumber: "12345678Z",
    }

    result, err := client.ValidateDocument(payload)

    if err != nil {
        fmt.Printf("Error de validación: %v\n", err)
        return
    }

    if result.Valid {
        fmt.Println("El documento es válido.")
    } else {
        fmt.Println("El documento no es válido.")
    }
}