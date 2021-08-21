package middlewares

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// This is should be the final middleware - cleans up all unhandled error and returns them as a response with status 500
type CleanUpMiddleware struct{}

func (errorMW CleanUpMiddleware) ModifyRequest(request Request) (Request, error) {
	return request, nil
}

func (errorMW CleanUpMiddleware) ModifyResponse(response Response, err error) (Response, error) {
	return response, err
}

func (errorMW CleanUpMiddleware) OnError(response Response, err error) (Response, error) {

	// if the second last middleware has handled the error, i.e. error == nil, just return the response
	if err == nil {
		return response, nil
	}

	objBytes, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		fmt.Println("cannot unmarshall response: " + marshalErr.Error())
	}
	stringBody := string(objBytes)
	stringBody = fmt.Sprintf("Unhandled error. Response received by final error handle middleware : '%s' ", stringBody)
	errorTrace := map[string]string{
		"stackTrace": errors.Wrap(err, stringBody).Error(),
	}

	objBytes, marshalErr = json.Marshal(errorTrace)
	if marshalErr != nil {
		fmt.Println("cannot unmarshall stackTrace:" + marshalErr.Error())
	}

	responseBody := string(objBytes)

	return Response{
		StatusCode:      500,
		IsBase64Encoded: false,
		Body:            responseBody,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}
