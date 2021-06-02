package request

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeQuery struct {
	ID   int    `url:"id"`
	Name string `url:"name"`
}

type fakeSuccess struct {
	ID   int `json:"ID"`
	Name string
}

type mockClient struct {
	mockHandler http.HandlerFunc
}

func (mtc *mockClient) Do(req *http.Request) (*http.Response, error) {
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mtc.mockHandler)
	handler.ServeHTTP(rr, req)

	return rr.Result(), nil
}

func fakeHandler(statusCode int, json string, headers map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(headers) > 0 {
			for key, value := range headers {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(statusCode)
		w.Write([]byte(json))
	}
}

func newMockRequest(fakeHandler http.HandlerFunc) *Request {
	return &Request{
		client: &mockClient{fakeHandler},
	}
}

func TestNew(t *testing.T) {
	request := New()

	t.Run("Nil", func(t *testing.T) {
		assert.NotNil(t, request)
	})

	t.Run("Method", func(t *testing.T) {
		assert.Equal(t, request.method, "GET")
	})

	t.Run("header type", func(t *testing.T) {
		assert.IsType(t, make(http.Header), request.header)
	})
}

func TestNewCopy(t *testing.T) {
	request := New()

	request.Post("http://example.com")
	request.SetQuery(&fakeQuery{ID: 20})
	request.AddHeader("Client-ID", "1234")
	request2 := request.New()

	expectedHeader := map[string][]string{"Client-Id": {"1234"}}

	t.Run("Not same object", func(t *testing.T) {
		//Must be a deep copy, not a pointer to the same object
		assert.NotSame(t, request, request2)
		assert.NotSame(t, request.header, request2.header)
		assert.NotSame(t, request.Failure, request2.Failure)

	})

	t.Run("deep equal", func(t *testing.T) {
		if !reflect.DeepEqual(request, request2) {
			t.Errorf("Objects are not deep equal expected %+v, got %+v", request, request2)
		}

		headerMap := map[string][]string(request2.header)

		if !reflect.DeepEqual(expectedHeader, headerMap) {
			t.Errorf("not equal: expected %v, got %v", expectedHeader, headerMap)
		}

	})
}

func TestSetSuccess(t *testing.T) {
	req := New()

	success := &fakeSuccess{ID: 2, Name: "John"}
	req.SetSuccess(success)
	assert.Equal(t, success, req.Success)
}

func TestSetFailure(t *testing.T) {
	req := New()

	failure := &fakeSuccess{ID: 2, Name: "John"}
	req.SetFailure(failure)
	assert.Equal(t, failure, req.Failure)
}

func TestAddHeader(t *testing.T) {
	cases := []struct {
		req            *Request
		expectedHeader map[string][]string
	}{
		{New().AddHeader("authorization", "Bearer 1234"), map[string][]string{"Authorization": {"Bearer 1234"}}},
		{New().AddHeader("content-tYPE", "application/json").AddHeader("User-AGENT", "chrome"), map[string][]string{"Content-Type": {"application/json"}, "User-Agent": {"chrome"}}},
		{New().AddHeader("A", "B").AddHeader("a", "c").New(), map[string][]string{"A": {"B", "c"}}},
		{New().AddHeader("A", "B").New().AddHeader("a", "c"), map[string][]string{"A": {"B", "c"}}},
	}
	for _, c := range cases {
		headerMap := map[string][]string(c.req.header)
		if !reflect.DeepEqual(c.expectedHeader, headerMap) {
			t.Errorf("not equal: expected %v, got %v", c.expectedHeader, headerMap)
		}
	}
}

func TestSetQuery(t *testing.T) {
	req := New()

	query := &fakeQuery{ID: 20, Name: "John"}
	req.SetQuery(query)

	request, err := req.Get("http://example.com").Request()

	expected := "http://example.com?id=20&name=John"
	assert.Nil(t, err)

	assert.Equal(t, request.URL.String(), expected)
}

func TestMethods(t *testing.T) {
	cases := []struct {
		req      *Request
		expected string
	}{
		{New().Get("http://example.com"), "GET"},
		{New().Post("http://example.com"), "POST"},
		{New().Delete("http://example.com"), "DELETE"},
		{New().Put("http://example.com"), "PUT"},
		{New().Patch("http://example.com"), "PATCH"},
		{New().Head("http://example.com"), "HEAD"},
	}

	for _, c := range cases {
		assert.Equal(t, c.req.method, c.expected)
	}
}

func TestSuccess(t *testing.T) {
	r := newMockRequest(fakeHandler(200, `{"id":200, "name":"John"}`, nil))

	result, err := r.Get("http://example.com").SetSuccess(&fakeSuccess{}).Execute()

	expected := &fakeSuccess{
		ID:   200,
		Name: "John",
	}
	assert.Nil(t, err)
	assert.Equal(t, result.StatusCode, 200)
	assert.Nil(t, result.Failure)
	assert.Equal(t, expected, result.Success.(*fakeSuccess))
}

func TestFailure(t *testing.T) {
	r := newMockRequest(fakeHandler(400, `{"id":200, "name":"John"}`, nil))

	result, err := r.Get("http://example.com").SetFailure(&fakeSuccess{}).Execute()

	expected := &fakeSuccess{
		ID:   200,
		Name: "John",
	}
	assert.Nil(t, err)
	assert.Equal(t, result.StatusCode, 400)
	assert.Nil(t, result.Success)
	assert.Equal(t, expected, result.Failure.(*fakeSuccess))
}

func TestSetBody(t *testing.T) {
	request := New()
	mockData := &fakeSuccess{ID: 10, Name: "Bob"}
	request.SetBody(mockData)

	req, err := request.Request()
	assert.Nil(t, err)

	body, err := json.Marshal(mockData)
	assert.Nil(t, err)

	buff := bytes.NewBuffer(body)
	bodyBytes, err := ioutil.ReadAll(req.Body)
	assert.Nil(t, err)

	assert.Equal(t, buff.Bytes(), bodyBytes)
}
