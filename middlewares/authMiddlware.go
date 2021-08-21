package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var errorMessage = "Unauthenticated, header Key value does not match"

// Performs authentication on request.
// middify := Middy{}
// middify.AddMiddleware(
//   &AuthMiddleWare{
// 	   KeyMap: map[string]string{
// 	     "Key": "123",
// }})
// wrappedHandler := middify.WrapHandler(Handler)
// curl -H "Key: 123" ...
type AuthMiddleWare struct {
	request Request
	KeyMap  map[string]string
}

func (authMW *AuthMiddleWare) ModifyRequest(request Request) (Request, error) {
	for key, value := range authMW.KeyMap {
		if request.Headers[key] != value {
			authMW.request = request
			fmt.Println("UNAUTHENTICATED")
			return request, errors.New("Unauthenticated, header Key value does not match")
		}
	}
	return request, nil
}

func (authMW *AuthMiddleWare) ModifyResponse(response Response, err error) (Response, error) {
	return response, err
}

func (authMW *AuthMiddleWare) OnError(response Response, _ error) (Response, error) {
	bytes, err := (json.MarshalIndent(authMW.request.Headers, "", "\t"))
	if err != nil {
		fmt.Println("cannot marshall authMW.request")
	}
	fmt.Println("Header:", string(bytes))

	bytes, err = (json.Marshal(map[string]string{
		"error": errorMessage,
	}))
	if err != nil {
		fmt.Println("cannot marshall httpError")
	}
	return Response{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		StatusCode: http.StatusUnauthorized,
		Body:       string(bytes),
	}, nil

}
