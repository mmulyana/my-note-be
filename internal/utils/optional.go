package utils

import "encoding/json"

// OptionalString dipakai di DTO update ketika perlu bedain field yang tidak dikirim (biarin apa adanya) dengan field yang dikirim null (hapus/clear nilainya di DB).
type OptionalString struct {
	Value   *string
	Present bool
}

func (o *OptionalString) UnmarshalJSON(data []byte) error {
	o.Present = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	o.Value = &s
	return nil
}
