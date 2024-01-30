package grep

import (
	"bytes"
	"errors"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestSearchString(t *testing.T) {
	dir1Name := "testDir"
	file1Name := "file1.txt"
	file2Name := "file2.txt"
	file3Name := "file3.txt"
	file4Name := "file4.txt"
	file1Data := []byte("this\nis\na\nfile\nIs")
	file2Data := []byte{}
	file3Data := []byte("line1\nline2\nline3\nline4\nline5\nline6 match1\nline7\nline8\nline9")
	file4Data := []byte("line1\nline2\nline3\nline4\nline5\nline6 match1\nline7 match2\nline8\nline9\nline10")

	testFS := fstest.MapFS{
		file1Name: {Data: file1Data, Mode: 0755},
		file2Name: {Data: file2Data, Mode: 0000},
		file3Name: {Data: file3Data, Mode: 0755},
		file4Name: {Data: file4Data, Mode: 0755},
		dir1Name:  {Data: nil, Mode: fs.ModeDir},
	}
	
	testCases := []struct {
		name       string
		fs         fs.FS
		stdin      []byte
		fileName   string
		keyword    string
		ignoreCase bool
		linesBeforeMatch int
		lineCount bool
		result     []string
		expErr     error
	}{
		{name: "greps a multi-line file", fs: testFS, stdin: nil, fileName: file1Name, keyword: "is", ignoreCase: false, result: []string{"this", "is"}, expErr: nil},
		{name: "greps a multi-line file text sensitive", fs: testFS, stdin: nil, fileName: file1Name, keyword: "is", ignoreCase: true, result: []string{"this", "is", "Is"}, expErr: nil},
		{name: "greps a multi-line file lines with before single match", fs: testFS, stdin: nil, fileName: file3Name, keyword: "match", ignoreCase: false, linesBeforeMatch: 2, result: []string{"line4", "line5", "line6 match1"}, expErr: nil},
		{name: "greps a multi-line file lines with before double match", fs: testFS, stdin: nil, fileName: file4Name, keyword: "match", ignoreCase: false, linesBeforeMatch: 2, result: []string{"line4", "line5", "line6 match1", "line5", "line6 match1", "line7 match2"}, expErr: nil},
		{name: "greps a multi-line file line with single count", fs: testFS, stdin: nil, fileName: file3Name, keyword: "match", ignoreCase: false, lineCount: true, result: []string{"1"}, expErr: nil},
		{name: "greps a multi-line file line with double count", fs: testFS, stdin: nil, fileName: file4Name, keyword: "match", ignoreCase: false, lineCount: true, result: []string{"2"}, expErr: nil},
		// {name: "greps a multi-line file lines after single match", fs: testFS, stdin: nil, fileName: file3Name, keyword: "match", ignoreCase: false, linesAfterMatch: 2, result: []string{"line6 match1", "line7", "line8"}, expErr: nil},
		{name: "reads from stdin", fs: nil, stdin: []byte("this\nis\na\nfile"), fileName: "", keyword: "is", result: []string{"this", "is"}, expErr: nil},
		{name: "reads a file with permission error", fs: testFS, stdin: nil, fileName: file2Name, expErr: fs.ErrPermission},
		{name: "reads an empty directory", fs: testFS, stdin: nil, fileName: dir1Name, expErr: ErrIsDirectory},
		{name: "reads a non-existent file", fs: testFS, stdin: nil, fileName: "non-existent-file.txt", expErr: fs.ErrNotExist},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := GrepOptions{Path: tc.fileName, Stdin: bytes.NewReader(tc.stdin), Keyword: tc.keyword, IgnoreCase: tc.ignoreCase, LinesBeforeMatch: tc.linesBeforeMatch, LineCount: tc.lineCount}
			got, err := Grep(tc.fs, options)
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

			if !reflect.DeepEqual(got, want) {
				t.Errorf("Expected %v but got %v", want, got)
			}
		})
	}
}

func TestSearchStringR(t *testing.T) {
	var testFS fstest.MapFS = make(map[string]*fstest.MapFile)
	testFS["tests"] = &fstest.MapFile{Data: nil, Mode: fs.ModeDir}
	testFS["tests/test1.txt"] = &fstest.MapFile{Data: []byte("Dummy Line\nthis is a test file\none can test a program by running test cases"), Mode: 0755}
	testFS["tests/filexyz.txt"] = &fstest.MapFile{Data: []byte("no matches here"), Mode: 0755}
	testFS["tests/inner/test1.txt"] = &fstest.MapFile{Data: []byte("dummy file"), Mode: 0755}
	testFS["tests/inner/test2.txt"] = &fstest.MapFile{Data: []byte("this file contains a test line"), Mode: 0755}

	testCases := []struct {
		name       string
		fs         fs.FS
		fileName   string
		keyword    string
		ignoreCase bool
		linesBeforeMatch int
		lineCount bool
		result     [][]string
		expErr     error
	}{
		{name: "greps inside a directory with -r", fs: testFS, fileName: "tests", keyword: "test", ignoreCase: false,result: [][]string{{"tests/test1.txt:this is a test file", "tests/test1.txt:one can test a program by running test cases"}, {"tests/inner/test2.txt:this file contains a test line"}}, expErr: nil},
		{name: "greps inside a directory with -r with lines before match option", fs: testFS, fileName: "tests", keyword: "test", ignoreCase: false, linesBeforeMatch: 1 ,result: [][]string{{"tests/test1.txt:Dummy Line", "tests/test1.txt:this is a test file","tests/test1.txt:this is a test file", "tests/test1.txt:one can test a program by running test cases"}, {"tests/inner/test2.txt:this file contains a test line"}}, expErr: nil},
		{name: "greps inside a directory with -r with line count option", fs: testFS, fileName: "tests", keyword: "test", ignoreCase: false, lineCount: true, result: [][]string{{"tests/test1.txt:2"}, {"tests/inner/test2.txt:1"}}, expErr: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := GrepOptions{Path: tc.fileName, Keyword: tc.keyword, IgnoreCase: tc.ignoreCase, LinesBeforeMatch: tc.linesBeforeMatch, LineCount: tc.lineCount}
			got, err := GrepR(tc.fs, options)
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

			// change the below checking of got & want into check of length (of sorts, maybe)
			// because no gurantee if order of elements will be as expected or not
			// also, if got length is zero, errors won't be catched
			for _, g := range got {
				matchFlag := false
				for _, w := range want {
					if reflect.DeepEqual(g, w) {
						matchFlag = true
						break
					}
				}

				if !matchFlag {
					t.Errorf("Expected %v but got %v", want, got)
				}
			}

			if len(got) != len(want) {
				t.Errorf("Expected length: %d, Got length: %d", len(want), len(got))
			}
		})
	}
}