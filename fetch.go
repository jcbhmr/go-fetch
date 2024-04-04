package fetch

import (
	"fmt"
	"regexp"

	"github.com/barweiss/go-tuple"
)

type HeadersInit interface {
	[][]string | map[string]string
}

type Headers struct {
	headerList map[string]string
	guardImmutable      bool
}

func NewHeaders[THeadersInit HeadersInit](init THeadersInit) *Headers {
	h := &Headers{
		headerList: map[string]string{},
		guardImmutable: false,
	}
	switch v := any(init).(type) {
	case nil: {}
	case [][]string:
		for _, tuple := range v {
			var name string
			if len(tuple) >= 1 {
				name = tuple[0]
			}
			var value string
			if len(tuple) >= 2 {
				value = tuple[1]
			}
			h.Set(name, value)
		}
	case map[string]string:
		for name, value := range v {
			h.Set(name, value)
		}
	default:
		panic(fmt.Errorf("unexpected type: %T", init))
	}
	return h
}

func (h *Headers) Append(name string, value string) {
	h.Set(name, value)
}

func (h *Headers) Delete(name string) {
	delete(h.headerList, name)
}

func (h *Headers) Get(name string) *string {
	value, ok := h.headerList[name]
	if ok {
		return &value
	} else {
		return nil
	}
}

func (h *Headers) Set(name string, value string) {
	h.headerList[name] = value
}

func (h *Headers) Iterable() map[string]string {
	iterable := map[string]string{}
	for name, value := range h.headerList {
		iterable[name] = value
	}
	return iterable
}

func Fetch() {}
