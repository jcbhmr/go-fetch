package fetch

import (
	"syscall/js"
)

type Headers struct {
	js.Value
}

var jsHeaders = js.Global().Get("Headers")

func NewHeaders() (*Headers, error) {
	jsValue := jsHeaders.New()
	return &Headers{jsValue}, nil
}
