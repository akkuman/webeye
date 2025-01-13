package utils

import (
	"fmt"
	"os"
	"testing"
)

func TestGetLocalFileOrWeb(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "example-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	content := "Hello, World!"
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tests := []struct{
		fileURL string
		content []byte
	} {
		{
			fileURL: "https://raw.githubusercontent.com/django/django/8bee7fa45cd7bfe70b68784314e994e2d193fd70/INSTALL",
			content: []byte(`Thanks for downloading Django.

To install it, make sure you have Python 3.10 or greater installed. Then run
this command from the command prompt:

    python -m pip install .

For more detailed instructions, see docs/intro/install.txt.
`),
		},
		{
			fileURL: tmpFile.Name(),
			content: []byte(content),
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("GetLocalFileOrWeb(%s)", tc.fileURL), func(t *testing.T) {
			content, err := GetLocalFileOrWeb(tc.fileURL)
			if err != nil {
				t.Error(err)
			}
			if string(content) != string(tc.content) {
				t.Errorf("got %s; want %s", string(content), string(tc.content))
			}
		})
	}
}
