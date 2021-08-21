package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

type Response = events.APIGatewayProxyResponse
type Request = events.APIGatewayProxyRequest
type Handler = func(request Request) (Response, error)

// Type of middlware function to process the request
type ModifyRequestFunc = func(request Request) (Request, error)

// Type of middleware function to process the response
type ModifyResponseFunc = func(response Response, err error) (Response, error)

// Type of middleware function to process errors return either by the middlewares, or by the main handler
type OnErrorFunc = func(response Response, err error) (Response, error)

// Interface for the middlewares
// Execution order of middlewares: https://github.com/middyjs/middy#execution-order
type MiddlewareImpl interface {
	// Middlware functions to process the request
	ModifyRequest(request Request) (Request, error)

	// Middleware functions to process the response
	ModifyResponse(response Response, err error) (Response, error)

	// Middlware functions to process errors returned either by the middlewares or the Handler function.
	// When there is an error, the regular control flow is stopped.
	// The execution is moved back to all the middlewares that implemented a special phase called onError, following the order they have been attached.
	// https://github.com/middyjs/middy#handling-errors

	// THE LAST ERROR HANDLING MIDDLEWARE SHOULD REMOVE THE ERROR, AND INSTEAD INCLUDE THE ERROR IN THE RESPONSE BODY, WITH STATUS 500, INDICATING UNHANDLED ERROR
	// The lambda handler can return 2 values. APIGatewayProxyResponse{} and error
	// The APIGatewayProxyResponse is simply a response, REGARDLESS of status code
	// If you return an error other then nil, then the Lambda failed with uncaught error with 502 Bad Gateway. https://stackoverflow.com/a/48462676/6514532
	OnError(response Response, err error) (Response, error)
}

// To use this middleware package
/*
	import (
		middy "github.com/serverless/sls-go-bulk/src/middleware"
	)

	middify := middy.Middy{}
	middify.AddMiddleware(middy.ErrorCleanUpMiddleware{})
	wrappedHandler := middify.WrapHandler(InjectedHandler)
	lambda.Start(wrappedHandler)
*/
type Middy struct {
	Middlewares []MiddlewareImpl
}

// Middify wraps the handler with the middlewares
func (middy *Middy) WrapHandler(handler Handler) Handler {
	return func(request Request) (Response, error) {
		// Logic to preprocess request here...
		var err error

		var modifyRequestFuncs []ModifyRequestFunc
		var modifyResponseFuncs []ModifyResponseFunc

		numMiddlewares := len(middy.Middlewares)
		for i := 0; i < numMiddlewares; i++ {
			modifyRequestMW := middy.Middlewares[i]
			modifyRequestFuncs = append(modifyRequestFuncs, modifyRequestMW.ModifyRequest)

			// modifyResponseFuncs need to be added in reverse order
			// https://github.com/middyjs/middy#execution-order
			reverseIndex := (numMiddlewares - 1) - i
			modifyResponseMW := middy.Middlewares[reverseIndex]
			modifyResponseFuncs = append(modifyResponseFuncs, modifyResponseMW.ModifyResponse)
		}

		for _, functn := range modifyRequestFuncs {
			request, err = functn(request)
			if err != nil {
				// since still in pre-processing request stage, provide an empty Response as initial value
				fmt.Printf("modify request middleware - error detected: %s \n", err.Error())
				return middy.handleError(Response{}, err)
			}
		}

		// Invoke main handler
		response, err := handler(request)
		if err != nil {
			return middy.handleError(response, err)
		}

		// Logic to process response here...
		for _, functn := range modifyResponseFuncs {
			response, err = functn(response, err)
			if err != nil {
				fmt.Printf("modify response middleware - error detected: %s \n", err.Error())
				return middy.handleError(response, err)
			}
		}

		return response, err
	}
}

func (middy *Middy) AddMiddleware(middleware ...MiddlewareImpl) {
	middy.Middlewares = append(middy.Middlewares, middleware...)
}

// Handle the error and create a proper response or to delegate the error to the next middleware.
func (middy *Middy) handleError(response Response, err error) (Response, error) {
	var onErrorFuncs []OnErrorFunc

	for _, mw := range middy.Middlewares {
		onErrorFuncs = append(onErrorFuncs, mw.OnError)
	}

	for _, functn := range onErrorFuncs {
		response, err = functn(response, err)
	}
	return response, err
}
