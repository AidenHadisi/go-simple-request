package request

import "net/http"

//Response is a response returned from the request
type Response struct {
	StatusCode int
	Header     http.Header
	Success    interface{}
	Failure    interface{}
}
