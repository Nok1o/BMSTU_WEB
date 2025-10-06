package forms

import (
	"encoding/json"
)

type PayloadWrapper[T any] struct {
	Payload T `json:"payload"`
}

func (pw *PayloadWrapper[T]) Unwrap() T {
	return pw.Payload
}

func (pw *PayloadWrapper[T]) MarshalJSON() ([]byte, error) {
	// Строим объект JSON с полем "payload"
	return json.Marshal(struct {
		Payload T `json:"payload"`
	}{
		Payload: pw.Payload,
	})
}
