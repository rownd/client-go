package main

import (
    "fmt"
    "log"
    
    "github.com/rgthelen/rownd-go-test/pkg/rownd/testing"
)

func main() {
    token, err := testing.GenerateTestToken()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Test token: %s\n", token)
}