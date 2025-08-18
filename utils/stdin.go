package utils

import (
	"bytes"
	"io"
	"os"
)

func ExpandStdin(text string) string {
	if text != "-" {
		return text
	}

	if stdin, err := io.ReadAll(os.Stdin); err == nil {
		return string(bytes.TrimSpace(stdin))
	}

	return ""
}
