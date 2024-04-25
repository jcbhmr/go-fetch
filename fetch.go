package fetch

import (
	"context"
	"io"
	"net/http"
	"net/url"
)



type conceptRequest struct {
	method     string
	url        *url.URL
	headerList map[string]string
	body       *io.Reader
	keepalive  bool
}

type Request struct {
	request *conceptRequest
	headers *Headers
	signal  *context.Context
	body    *io.Reader
}

func NewRequest(input string, init *RequestInit) *Request {
	var headers *Headers
	if init != nil && init.Headers != nil {
		headers = init.Headers
	} else {
		headers = NewHeaders(nil)
	}
	url, err := url.Parse(input)
	if err != nil {
		panic(err)
	}
	var method string
	if init != nil && init.Method != nil {
		method = *init.Method
	} else {
		method = "GET"
	}
	return &Request{
		request: &conceptRequest{
			method:     method,
			url:        url,
			headerList: headers.headerList,
			body:       nil,
			keepalive:  false,
		},
		headers: headers,
		signal:  nil,
		body:    nil,
	}
}

type RequestInit struct {
	Method         *string
	Headers        *Headers
	Body           **string
	Referrer       *string
	ReferrerPolicy *string
	Mode           *string
	Credentials    *string
	Cache          *string
	Redirect       *string
	Integrity      *string
	Keepalive      *bool
	Signal         *context.Context
	Duplex         *string
	Priority       *string
	Window         *any
}

type Response struct{}

type FetchResult struct {
	*Response
	Err error
}

func Fetch(input string, init *RequestInit) <-chan FetchResult {
	c := make(chan FetchResult)
	go func() {
		defer close(c)
		res, err := http.Get(input)
		if err != nil {
			c <- FetchResult{Response: nil, Err: err}
			return
		}
		c <- FetchResult{Response: &Response{}, Err: nil}
	}()
	return c
}
