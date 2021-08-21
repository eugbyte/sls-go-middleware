package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/eugbyte/sls-go-mod/middlewares"
)

// Test the basic wrapping method
func TestMiddify(t *testing.T) {
	middify := Middy{}

	wrappedHandler := middify.WrapHandler(MockHandler)
	mockRequest := Request{}
	_, err := wrappedHandler(mockRequest)
	if err != nil {
		t.Errorf("Expected no error, but received %v", err)
	}
}

func MockHandler(request Request) (Response, error) {
	return Response{}, nil
}

// Test the ErrorCleanUpMiddleware
// By default, middy does not clean
func TestErrorClean(t *testing.T) {
	middify := Middy{}
	middify.AddMiddleware(middlewares.CleanUpMiddleware{})

	wrappedHandler := middify.WrapHandler(MockErrorHandler)
	mockRequest := Request{}

	t.Log("Without the CleanUpMiddleware, the error from a handler that returns (Response, error) should remain")
	// Without the CleanUpMiddleware, error should remain
	_, err := MockErrorHandler(mockRequest)
	if err != nil {
		t.Logf("Test passed, expected error, and error recevied: %v", err)
	} else {
		t.Errorf("Expected error, but no error received.")
	}

	t.Log("Without the CleanUpMiddleware, the error from a handler that returns (Response, error) should be nil")
	_, err = wrappedHandler(mockRequest)

	if err != nil {
		t.Errorf("Expected no error, but error received %v.", err)
	} else {
		t.Logf("Test passed, no error")
	}

}

func MockErrorHandler(request Request) (Response, error) {
	response := Response{}
	response.Body = "eee"
	return response, errors.New("Mock error")
}

func TestMiddlewareFlow(t *testing.T) {
	middify := Middy{}
	middify.AddMiddleware(
		&middlewares.AuthMiddleWare{
			KeyMap: map[string]string{
				"Key": "123",
			}},
		middlewares.CleanUpMiddleware{})
	wrappedHandler := middify.WrapHandler(MockHandler)
	mockRequest := Request{
		Headers: map[string]string{"Key": "Wrong key"},
	}
	response, err := wrappedHandler(mockRequest)
	t.Log("The api key does not match => AuthMiddleware will return response with 403, and the error described in the response body")
	t.Log("The final cleanup middleware should not be triggered, because authMiddleware handles the error by returning nil error, and thus no clean up is needed")
	if err != nil {
		t.Errorf("Expected no error, but received %v", err)
	}

	if response.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code to be %d, but received %d", http.StatusUnauthorized, response.StatusCode)
	}

	var errorBody map[string]string
	err = json.Unmarshal([]byte(response.Body), &errorBody)
	if err != nil {
		t.Errorf("Could not unmarshall response.Body")
	}

	expectedErrorMessage := "Unauthenticated, header Key value does not match"
	actualErrorMessage := errorBody["error"]
	if actualErrorMessage != expectedErrorMessage {
		t.Errorf("Expected error message to be %s, but received %s", expectedErrorMessage, actualErrorMessage)
	}

}
