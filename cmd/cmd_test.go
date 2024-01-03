package cmd

import (
	"errors"
	"os"
	"testing"
)

func TestWriteToFile(t *testing.T) {
	testCases := []struct{
		name string
		filePath string
		content string
		expErr error
	}{
		{name: "write to file", filePath: "test.txt", content: "test only", expErr: nil},
		{name: "write to already created file", filePath: "test.txt", content: "test only", expErr: os.ErrExist},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expErr != nil {
				os.Create(tc.filePath)
			}

			err := writeToFile(tc.filePath, tc.content)
			defer os.Remove(tc.filePath)

			if tc.expErr != nil {
				if err == nil {
					t.Fatalf("Expected error but didn't got one")
				}

				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error %v but got %v", tc.expErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			data, err := os.ReadFile(tc.filePath)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if string(data) != tc.content {
				t.Errorf("Expected %q but got %q", string(data), tc.content)
			}
		})
	}
}