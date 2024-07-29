package util

import (
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func DecodeAll(path string) ([]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	decoder := yaml.NewDecoder(f)
	var result []any
	for {
		var data any
		if err := decoder.Decode(&data); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		result = append(result, data)
	}

	return result, nil
}
