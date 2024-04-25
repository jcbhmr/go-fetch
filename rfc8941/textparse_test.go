package rfc8941

import (
	"encoding/json"
	"testing"
)

func ptr[T any](v T) *T {
	return &v
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// https://github.com/httpwg/structured-field-tests
var textParseTests = []struct {
	InputBytes []byte
	FieldType  string
	MustFail   bool
}{
	{[]byte(`text/html;q=1.0`), "list", false},
	{[]byte(`text/html  ,  text/plain;  q=0.5;  charset=utf-8`), "list", false},
	{[]byte(`a=1, b;foo=9, c=3`), "dictionary", false},
}

func TestTextParse(t *testing.T) {
	for _, tt := range textParseTests {
		t.Run(string(tt.InputBytes), func(t *testing.T) {
			t.Logf("InputBytes=%q", tt.InputBytes)
			t.Logf("FieldType=%#+v", tt.FieldType)
			t.Logf("MustFail=%#+v", tt.MustFail)
			value, err := TextParse(tt.InputBytes, tt.FieldType)
			t.Logf("value=%s", mustMarshal(value))
			t.Logf("err=%#+v", err)
			if tt.MustFail && err == nil {
				t.Errorf("expected an error")
			} else if !tt.MustFail && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
