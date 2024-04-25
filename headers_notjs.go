//go:build !js

package fetch

import (
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/jcbhmr/go-fetch/rfc8941"
)

// A header list is a list of zero or more headers. It is initially « ».
//
// A header list is essentially a specialized multimap: an ordered list of
// key-value pairs with potentially duplicate keys. Since headers other than
// Set-Cookie are always combined when exposed to client-side JavaScript,
// implementations could choose a more efficient representation, as long as they
// also support an associated data structure for Set-Cookie headers.
//
// https://fetch.spec.whatwg.org/#concept-header-list
type headerList []header2

// To get a structured field value given a header name name and a string type
// from a header list list, run these steps. They return null or a structured
// field value.
func (h headerList) GetStructuredHeader(name headerName, type_ string) rfc8941.StructuredFieldValue {
	// 1. Assert: type is one of "dictionary", "list", or "item".
	if !slices.Contains([]string{"dictionary", "list", "item"}, type_) {
		panic(fmt.Errorf(`type is not "dictionary"|"list"|"item". got %#v`, type_))
	}

	// 2. Let value be the result of getting name from list.
	value := h.Get(name)

	// 3. If value is null, then return null.
	if value == nil {
		return nil
	}

	// 4. Let result be the result of parsing structured fields with input_string set to value and header_type set to type.
	result, err := rfc8941.TextParse(value, type_)
	// 5. If parsing failed, then return null.
	if err != nil {
		return nil
	}
	// 6. Return result.
	return result
}

// 
type header2 struct {
	Name  conceptHeaderName
	Value conceptHeaderValue
}

// https://fetch.spec.whatwg.org/#header-name
type conceptHeaderName []byte

var fieldName = regexp.MustCompile(`^[!#$%&'*+\-\.^_\` + "`" + `|~0-9A-Za-z]+$`)

func newConceptHeaderName(name []byte) (conceptHeaderName, error) {
	if fieldName.Match(name) {
		return conceptHeaderName(name), nil
	} else {
		return nil, errors.New("did not match fieldName")
	}
}

// https://fetch.spec.whatwg.org/#header-value
type conceptHeaderValue []byte

var leadingHTTPTabOrSpaceBytes = regexp.MustCompile(`^[\t ]+`)
var trailingHTTPTabOrSpaceBytes = regexp.MustCompile(`[\t ]+$`)
var nulOrHTTPNewlineBytes = regexp.MustCompile(`[\x00\x0A\x0D]`)

func newConceptHeaderValue(value []byte) (conceptHeaderValue, error) {
	if leadingHTTPTabOrSpaceBytes.Match(value) {
		return nil, errors.New("matched leadingHTTPTabOrSpaceBytes")
	} else if trailingHTTPTabOrSpaceBytes.Match(value) {
		return nil, errors.New("matched trailingHTTPTabOrSpaceBytes")
	} else if nulOrHTTPNewlineBytes.Match(value) {
		return nil, errors.New("matched nulOrHTTPNewlineBytes")
	} else {
		return conceptHeaderValue(value), nil
	}
}

// https://fetch.spec.whatwg.org/#concept-header-value-normalize
func normalize(potentialValue []byte) []byte {
	v := potentialValue
	v = leadingHTTPTabOrSpaceBytes.ReplaceAll(v, []byte{})
	v = trailingHTTPTabOrSpaceBytes.ReplaceAll(v, []byte{})
	return v
}

// https://fetch.spec.whatwg.org/#typedefdef-headersinit
type HeadersInit interface {
	[][]string|map[string]string
}

// https://fetch.spec.whatwg.org/#headers-class
type Headers struct {
	headerList     conceptHeaderList
	guard headersGuard
}

// https://fetch.spec.whatwg.org/#headers-guard
type headersGuard string

func validate(nameValue struct{Name string;Value string}, headers *Headers) (bool, error) {
	name := nameValue.Name
	value := nameValue.Value
	if headerName, err := newConceptHeaderName([]byte(name)); err != nil {
		return false, err
	}
	if headerValue, err := newConceptHeaderValue([]byte(value)); err != nil {
		return false, err
	}
	if headers.guard == "immutable" {
		return false, errors.New("guard is immutable")
	}
	if headers.guard == "request" && isForbiddenRequestHeader(struct{Name string;Value string}{name, value}) {
		return false, errors.New("forbidden request header")
	}
}

func NewHeaders[T HeadersInit](init T) *Headers {
	return nil
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
