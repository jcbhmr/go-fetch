package rfc8941_test

import (
	"fmt"
	"testing"

	"github.com/jcbhmr/go-fetch/internal/rfc8941"
	"github.com/samber/lo"
)

var textSerializeTests = []struct {
	InputValue any
	MustFail   bool
}{
	{rfc8941.List{}, false},
	{rfc8941.Dictionary{}, false},
	{rfc8941.Item{A:"text/html",B:rfc8941.Parameters{lo.T2("q", any(1.0))}}, false},
}

func TestTextSerialize(t *testing.T) {
	for _, tt := range textSerializeTests {
		t.Run(fmt.Sprint(tt.InputValue), func(t *testing.T) {
			t.Logf("InputValue=%#+v", tt.InputValue)
			t.Logf("MustFail=%#+v", tt.MustFail)
			value, err := rfc8941.TextSerialize(tt.InputValue)
			t.Logf("value=%s", value)
			t.Logf("err=%#+v", err)
			if tt.MustFail && err == nil {
				t.Errorf("expected an error")
			} else if !tt.MustFail && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
