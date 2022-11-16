package helper

import (
	"fmt"
	"testing"
)

func Test_stringWithCharset(t *testing.T) {
	tests := []struct {
		name    string
		length  int64
		charset string
	}{
		{
			name:    "normal string",
			length:  32,
			charset: charset,
		},
		{
			name:    "0 string",
			length:  0,
			charset: charset,
		},
		{
			name:    "0 charset",
			length:  32,
			charset: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := stringWithCharset(tt.length, tt.charset)
			fmt.Printf("%v\n", str)
			if err != nil {
				t.Error("stringWithCharset with errors")
			}
			if len(str) != int(tt.length) {
				t.Errorf("Length of string is not as eypexted: %v != %v", len(str), tt.length)
			}
		})
	}
}
