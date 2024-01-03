package grep

import (
	"bytes"
	"errors"
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"
)

func TestSearchString(t *testing.T) {
	file1Name := "file1.txt"
	file2Name := "file2.txt"
	file3Name := "testDir"
	file4Name := "non-existent.txt"

	file1Data := []byte("this\nis\na\nfile")
	file2Data := []byte{}

	testFS := fstest.MapFS{
		file1Name: {Data: file1Data, Mode: 0755},
		file2Name: {Data: file2Data, Mode: 0000},
		file3Name: {Data: nil, Mode: fs.ModeDir},
	}

	testCases := []struct{
		name string
		fs fs.FS
		stdin []byte
		fileName string
		keyword string
		result []string
		expErr error
	}{
		{name: "reads a normal multi-line file", fs: testFS, stdin: nil, fileName: file1Name, keyword: "is", result: []string{"this", "is"}, expErr: nil},
		{name: "reads from stdin", fs: nil, stdin: []byte("this\nis\na\nfile"), fileName: "", keyword: "is", result: []string{"this", "is"}, expErr: nil},
		{name: "reads a file with permission error", fs: testFS, stdin: nil, fileName: file2Name, expErr: fs.ErrPermission},
		{name: "reads an empty directory", fs: testFS, stdin: nil, fileName: file3Name, expErr: ErrIsDirectory},
		{name: "reads a non-existent file", fs: testFS, stdin: nil, fileName: file4Name, expErr: fs.ErrNotExist},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SearchString(tc.fs, tc.fileName, bytes.NewReader(tc.stdin), tc.keyword)
			want := tc.result

			if tc.expErr != nil {
				if err == nil {
					t.Fatalf("Expected an error but didn't got one")
				}

				if !errors.Is(err, tc.expErr) {
					t.Fatalf("Expected error %v but got %v", tc.expErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Didn't expected an error: %v", err)
			}

			if !slices.Equal(got, want) {
				t.Errorf("Expected %v but got %v", want, got)
			}
		})
	}
}