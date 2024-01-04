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
	dir1Name := "testDir"
	file4Name := "non-existent.txt"
	// dir2Name := "tests"

	file1Data := []byte("this\nis\na\nfile\nIs")
	file2Data := []byte{}

	testFS := fstest.MapFS{
		file1Name: {Data: file1Data, Mode: 0755},
		file2Name: {Data: file2Data, Mode: 0000},
		dir1Name: {Data: nil, Mode: fs.ModeDir},
		
		// "tests/test1.txt": {Data: []byte("Dummy Line\nthis is a test file\none can test a program by running test cases"), Mode: 0755},
		// "tests/filexyz.txt" : {Data: []byte("no matches here"), Mode: 0755},
    	// "tests/inner/test1.txt" : {Data: []byte("dummy file"), Mode: 0755},
    	// "tests/inner/test2.txt" : {Data: []byte("this file contains a test line"), Mode: 0755},
	}

	// testFS["tests/test1.txt"] = &fstest.MapFile{Data: []byte("Dummy Line\nthis is a test file\none can test a program by running test cases"), Mode: 0755}
    // testFS["tests/filexyz.txt"] = &fstest.MapFile{Data: []byte("no matches here"), Mode: 0755}
    // testFS["tests/inner/test1.txt"] = &fstest.MapFile{Data: []byte("dummy file"), Mode: 0755}
    // testFS["tests/inner/test2.txt"] = &fstest.MapFile{Data: []byte("this file contains a test line"), Mode: 0755}

	testCases := []struct{
		name string
		fs fs.FS
		stdin []byte
		fileName string
		keyword string
		ignoreCase bool
		searchDir bool
		result []string
		expErr error
	}{
		{name: "greps a normal multi-line file", fs: testFS, stdin: nil, fileName: file1Name, keyword: "is", ignoreCase: false, result: []string{"this", "is"}, expErr: nil},
		{name: "greps a normal multi-line file text sensitive", fs: testFS, stdin: nil, fileName: file1Name, keyword: "is", ignoreCase: true, result: []string{"this", "is", "Is"}, expErr: nil},
		{name: "reads from stdin", fs: nil, stdin: []byte("this\nis\na\nfile"), fileName: "", keyword: "is", result: []string{"this", "is"}, expErr: nil},
		{name: "reads a file with permission error", fs: testFS, stdin: nil, fileName: file2Name, expErr: fs.ErrPermission},
		{name: "reads an empty directory", fs: testFS, stdin: nil, fileName: dir1Name, expErr: ErrIsDirectory},
		{name: "reads a non-existent file", fs: testFS, stdin: nil, fileName: file4Name, expErr: fs.ErrNotExist},
		// {name: "greps inside a directory with -r", fs: testFS, stdin: nil, fileName: dir2Name, keyword: "tests", ignoreCase: false, searchDir: true, expErr: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SearchString(tc.fs, tc.fileName, bytes.NewReader(tc.stdin), tc.keyword, tc.ignoreCase, tc.searchDir)
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

func TestSearchStringR(t *testing.T) {
	var testFS fstest.MapFS = make(map[string]*fstest.MapFile)
	testFS["tests"] = &fstest.MapFile{Data: nil}
	testFS["tests/test1.txt"] = &fstest.MapFile{Data: []byte("Dummy Line\nthis is a test file\none can test a program by running test cases")}
    testFS["tests/filexyz.txt"] = &fstest.MapFile{Data: []byte("no matches here")}
    testFS["tests/inner/test1.txt"] = &fstest.MapFile{Data: []byte("dummy file")}
    testFS["tests/inner/test2.txt"] = &fstest.MapFile{Data: []byte("this file contains a test line")}

	testCases := []struct{
		name string
		fs fs.FS
		stdin []byte
		fileName string
		keyword string
		ignoreCase bool
		searchDir bool
		result []string
		expErr error
	}{
		{name: "greps inside a directory with -r", fs: testFS, stdin: nil, fileName: "tests", keyword: "tests", ignoreCase: false, result: []string{"this is a test file", "one can test a program by running test cases", "this file contains a test line"}, searchDir: true, expErr: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SearchString(tc.fs, tc.fileName, bytes.NewReader(tc.stdin), tc.keyword, tc.ignoreCase, tc.searchDir)
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