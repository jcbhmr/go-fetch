package fetch_test

import (
	"fmt"

	"github.com/jcbhmr/go-fetch"
)

func ExampleNewHeaders() {
	headers := fetch.NewHeaders(nil)
	for name, value := range headers.Iterable() {
		fmt.Printf("%#+v: %#+v\n", name, value)
	}
}
