// Package validate creates a shared validator instance. Use this to parse and validate structs.
package validate

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
)

var V = validator.New(validator.WithRequiredStructEnabled())

func ParseJSON[T any](payload io.Reader) (*T, error) {
	parsed := new(T)
	decoder := json.NewDecoder(payload)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(parsed); err != nil {
		return nil, fmt.Errorf("decode strict json: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode strict json: unexpected trailing data")
		}
		return nil, fmt.Errorf("decode strict json: %w", err)
	}

	if err := V.Struct(parsed); err != nil {
		return nil, fmt.Errorf("validate strict json: %w", err)
	}

	return parsed, nil
}
