package main

import (
    "fmt",
    "github.com/Luhnify/luhnify-go"
)

func main() {
    apiKey := "YOUR_API_KEY_HERE"

    client := validation.NewValidationClient(apiKey, nil)

    payload := validation.ValidationPayload{
        CountryCode:    "es",
        DocumentType:   "dni",
        DocumentNumber: "12345678Z",
    }

    result, err := client.ValidateDocument(payload)

    if err != nil {
        fmt.Printf("Validation error: %v\n", err)
        return
    }

    if result.Valid {
        fmt.Println("Success: The document is valid.")
    } else {
        fmt.Println("Warning: The document is invalid.")
    }
}