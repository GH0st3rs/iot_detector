package main

import (
	"encoding/json"
	"os"
)

type Request struct {
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Search  string            `json:"search"`
	Data    string            `json:"data"`
}

func (r *Request) loadRequestFromFile(filename string) error {
	// Read file
	if byteValue, err := os.ReadFile(filename); err != nil {
		return err
		// Parse Json file
	} else if err := json.Unmarshal(byteValue, r); err != nil {
		return err
	}
	return nil
}
