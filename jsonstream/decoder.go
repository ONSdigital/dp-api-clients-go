package jsonstream

import (
	"encoding/json"
	"fmt"
	"io"
)

// Decoder is a json.Decoder wrapper which adds convenience
// methods for stream decoding and uses panic to simplify errors.
// You should use recover() to catch errors from the methods.
type Decoder struct{ *json.Decoder }

// New creates a new Decoder.
// Decoder is a pointer type: copying does not clone state.
func New(r io.Reader) Decoder {
	jd := json.NewDecoder(r)
	jd.UseNumber()
	return Decoder{jd}
}

// StartObjectComposite decodes the start of a JSON object, i.e. '{'
func (dec Decoder) StartObjectComposite() bool { return dec.start('{') }

// StartArrayComposite decodes the start of a JSON array, i.e. '['
func (dec Decoder) StartArrayComposite() bool { return dec.start('[') }

// start array or object
func (dec Decoder) start(delim json.Delim) bool {
	tok := dec.mustToken()
	if tok == nil {
		return false
	}
	gotDelim, ok := tok.(json.Delim)
	if !ok {
		panic(fmt.Sprintf("Expected %q but got %q", delim, tok))
	}
	if gotDelim != delim {
		panic(fmt.Sprintf("Expected %q but got %q", delim, gotDelim))
	}
	return true
}

// EndComposite will decode and discard the end of an array or object
func (dec Decoder) EndComposite() {
	// json.Decoder guarantees matching delimiters so no need to check
	_ = dec.mustToken()
}

// DecodeString decodes a token and check that it is a string or null.
// It returns nil if a null was found.
func (dec Decoder) DecodeString() *string {
	tok := dec.mustToken()
	if tok == nil {
		return nil
	}
	s, ok := tok.(string)
	if !ok {
		panic(fmt.Sprintf("Expected string but got %q", tok))
	}
	return &s
}

// DecodeName decodes a token and checks that it is a non-null string
func (dec Decoder) DecodeName() string {
	if s := dec.DecodeString(); s != nil {
		return *s
	}
	panic("Expected JSON field name but got null")
}

// DecodeNumber decodes a token and checks that it is a non-null number
func (dec Decoder) DecodeNumber() json.Number {
	tok := dec.mustToken()
	n, ok := tok.(json.Number)
	if !ok {
		panic(fmt.Sprintf("Expected number but got %q", tok))
	}
	return n
}

// mustToken reads a token and panics on error
func (dec Decoder) mustToken() json.Token {
	tok, err := dec.Token()
	if err != nil {
		panic(err)
	}
	return tok
}
