package jsonstream

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Decoder is a json.Decoder wrapper which adds convenience methods for stream decoding.
type Decoder struct{ *json.Decoder }

// New creates a new Decoder.
// Decoder is a pointer type: copying does not clone state.
func New(r io.Reader) Decoder {
	jd := json.NewDecoder(r)
	jd.UseNumber()
	return Decoder{jd}
}

// StartObjectComposite decodes the start of a JSON object, i.e. '{'
func (dec Decoder) StartObjectComposite() (bool, error) { return dec.start('{') }

// StartArrayComposite decodes the start of a JSON array, i.e. '['
func (dec Decoder) StartArrayComposite() (bool, error) { return dec.start('[') }

// start array or object
func (dec Decoder) start(delim json.Delim) (bool, error) {
	tok, err := dec.Token()
	if err != nil {
		return false, fmt.Errorf("error decoding json token: %w", err)
	}

	if tok == nil {
		return false, nil
	}

	gotDelim, ok := tok.(json.Delim)
	if !ok {
		return false, fmt.Errorf("expected %q but got %q", delim, tok)
	}

	if gotDelim != delim {
		return false, fmt.Errorf("expected %q but got %q", delim, gotDelim)
	}
	return true, nil
}

// EndComposite will decode and discard the end of an array or object
func (dec Decoder) EndComposite() error {
	// json.Decoder guarantees matching delimiters so no need to check
	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("error decoding json token: %w", err)
	}
	return nil
}

// DecodeString decodes a token and check that it is a string or null.
// It returns nil if a null was found.
func (dec Decoder) DecodeString() (*string, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("error decoding json token: %w", err)
	}

	if tok == nil {
		return nil, nil
	}

	s, ok := tok.(string)
	if !ok {
		return nil, fmt.Errorf("expected string but got %q", tok)
	}

	return &s, nil
}

// DecodeName decodes a token and checks that it is a non-null string
func (dec Decoder) DecodeName() (string, error) {
	s, err := dec.DecodeString()
	if err != nil {
		return "", fmt.Errorf("error decoding string: %w", err)
	}

	if s == nil {
		return "", errors.New("expected json field name but got null")
	}

	return *s, nil
}

// DecodeNumber decodes a token and checks that it is a non-null number
func (dec Decoder) DecodeNumber() (json.Number, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", fmt.Errorf("error decoding json token: %w", err)
	}

	n, ok := tok.(json.Number)
	if !ok {
		return "", fmt.Errorf("expected number but got %q", tok)
	}

	return n, nil
}
