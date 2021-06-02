//Pacakge request provides a http client
//which includes everything you need for simple requests
package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"time"

	"github.com/google/go-querystring/query"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

//Request is a simple http request client
type Request struct {
	client  httpClient
	method  string
	url     string
	header  http.Header
	query   interface{}
	body    interface{}
	Success interface{}
	Failure interface{}
}

//New creates a new Request
func New() *Request {
	return &Request{
		client: &http.Client{Timeout: time.Second * 3},
		method: "GET",
		header: make(http.Header),
	}
}

//New creates a new request from existing request
func (r *Request) New() *Request {
	headers := make(http.Header)
	for key, value := range r.header {
		headers[key] = value
	}

	return &Request{
		client:  r.client,
		method:  r.method,
		url:     r.url,
		header:  headers,
		query:   r.query,
		body:    r.body,
		Success: r.Success,
		Failure: r.Failure,
	}
}

//SetSuccess is used to set a custom struct for response body unmarshalling after a successful request
//Must be passed as a reference
func (r *Request) SetSuccess(success interface{}) *Request {
	r.Success = success
	return r
}

//SetFailure is used to set a custom struct for response body unmarshalling after a failed request
//Must be passed as a reference
func (r *Request) SetFailure(failure interface{}) *Request {
	r.Failure = failure
	return r
}

//SetHeader can be used to set a header for the request
func (r *Request) SetHeader(key, value string) *Request {
	r.header.Set(key, value)
	return r
}

//AddHeader can be used to add a header for the request
func (r *Request) AddHeader(key, value string) *Request {
	r.header.Add(key, value)
	return r
}

//SetQuery is used to set query params for request
func (r *Request) SetQuery(query interface{}) *Request {
	r.query = query
	return r
}

//SetBody is used to set request body. Must be passed as a pointer to a struct
func (r *Request) SetBody(body interface{}) *Request {
	r.body = body
	return r
}

//Get request
func (r *Request) Get(url string) *Request {
	r.method = "GET"
	return r.setURL(url)

}

//Post request
func (r *Request) Post(url string) *Request {
	r.method = "POST"
	return r.setURL(url)

}

//Put request
func (r *Request) Put(url string) *Request {
	r.method = "PUT"
	return r.setURL(url)

}

//Head request
func (r *Request) Head(url string) *Request {
	r.method = "HEAD"
	return r.setURL(url)

}

//Delete request
func (r *Request) Delete(url string) *Request {
	r.method = "DELETE"
	return r.setURL(url)

}

//Patch request
func (r *Request) Patch(url string) *Request {
	r.method = "PATCH"
	return r.setURL(url)

}

//Request creates and returns and http request
func (r *Request) Request() (*http.Request, error) {
	var req *http.Request
	var err error

	if r.body != nil {
		body, err := json.Marshal(r.body)
		if err != nil {
			return nil, err
		}
		buff := bytes.NewBuffer(body)

		req, err = http.NewRequest(r.method, r.url, buff)
	} else {
		req, err = http.NewRequest(r.method, r.url, nil)
	}

	if err != nil {
		return nil, err
	}

	v, err := query.Values(r.query)
	if err == nil {
		req.URL.RawQuery = v.Encode()
	}

	return req, nil
}

//Execute runs the request and returns a response
func (r *Request) Execute() (*Response, error) {
	return r.sendRequest()
}

func (r *Request) setURL(address string) *Request {
	path, err := url.Parse(address)
	if err == nil {
		r.url = path.String()
	}

	return r
}

func (r *Request) sendRequest() (*Response, error) {

	req, err := r.Request()
	if err != nil {
		return nil, err
	}
	resp, err := r.do(req)

	return resp, err
}

func (r *Request) do(req *http.Request) (*Response, error) {

	response := &Response{}
	resp, err := r.client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// if resp.StatusCode == http.StatusNoContent || resp.ContentLength == 0 {
	// 	return response, nil
	// }

	response.StatusCode = resp.StatusCode
	response.Header = resp.Header

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = r.decodeResp(response, bodyBytes)

	if err != nil {
		err = fmt.Errorf("failed to decode API response: %s", err.Error())
	}

	return response, err
}

func (r *Request) decodeResp(resp *Response, body []byte) error {
	if status := resp.StatusCode; 200 <= status && status <= 299 {
		if r.Success != nil {
			resp.Success = r.Success

			return json.Unmarshal(body, &resp.Success)
		}

	} else {
		if r.Failure != nil {
			resp.Failure = r.Failure
			return json.Unmarshal(body, &resp.Failure)
		}
	}
	return nil

}
