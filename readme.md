## Middleware for serverless / aws lambda in go lang
Inspired by [middy](https://github.com/middyjs/middy)
Note that this is not the official release of middy

# How to use
```
import (
    "github.com/aws/aws-lambda-go/lambda"
    middy "github.com/eugbyte/sls-go-middleware"

)

middify := middy.Middy{}
middify.AddMiddleware(
    middlware1,
    middlware2,
)

wrappedHandler := middify.WrapHandler(Handler)
lambda.Start(wrappedHandler)
```
# How to create your own middleware
All middleware must implement the `MiddlewareImpl` interface
```
// Interface for the middlewares
type MiddlewareImpl interface {
	// Middlware functions to process the request
	ModifyRequest(request Request) (Request, error)

	// Middleware functions to process the response
	ModifyResponse(response Response, err error) (Response, error)

	// Middlware functions to process errors returned either by the middlewares or the Handler function.
	OnError(response Response, err error) (Response, error)
}
```

# Understanding how it works
* [middy wraps the handler in the middlewares specified](https://github.com/middyjs/middy#how-it-works)

* [Execution order](https://github.com/middyjs/middy#execution-order)
```
## The ModifyRequest functions will be executed in the order from middleware1, middleware2, middleware3
## The ModifyResponse functions will be executed in the order from middleware3, middleware2, middleware1
## The OnError functions will be executed in the order from middleware1, middleware2, middleware3

middify.AddMiddleware(
    middleware1,
    middleware2,
    middleware3,
)
```


# Error handling
## Understanding the pecularities of the returning an error in aws lambda
The lambda handler can return 2 values. APIGatewayProxyResponse{} and error.

The APIGatewayProxyResponse is simply a response, REGARDLESS of status code.

If you return an error other then nil, then the Lambda failed with uncaught error and with 502 Bad Gateway. https://stackoverflow.com/a/48462676/6514532

Thus, the last error handling middleware should remove the error, and instead include the error in the response body, with status 500, indicating unhandled error.

The library already provides a cleanUp middleware to cleans up all unhandled error and returns them as a response with status 500. This middleware should be used as the final middleware, e.g.:
```
import (
	middy "github.com/eugbyte/sls-go-mod"
)

middify := middy.Middy{}
middify.AddMiddleware(
    middlware1,
    middlware2,
    middy.middlewares.CleanUpMiddleware{},
)

wrappedHandler := middify.WrapHandler(Handler)
```

# Testing the library
// Install make (win | linux)

`choco install make` | `apt-get install make`

// Install gotest

`make test-install-gotest`

// Run the tests

`make test`
