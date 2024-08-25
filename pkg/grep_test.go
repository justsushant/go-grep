package grep

import (
	"bytes"
	"errors"
	"io/fs"
	"reflect"
	"slices"
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
		name             string
		stdin            []byte
		fileName         string
		keyword          string
		ignoreCase       bool
		linesBeforeMatch int
		linesAfterMatch  int
		lineCount        bool
		result           GrepResult
		expErr           error
	}{
		{
			name:       "greps a multi-line file",
			fileName:   file1Name,
			keyword:    "is",
			ignoreCase: false,
			result:     GrepResult{MatchedLines: []string{"this", "is"}},
			// result:     GrepResult{Path: file1Name, MatchedLines: []string{"this", "is"}},
			expErr:     nil,
		},
		{
			name:       "greps a multi-line file text sensitive",
			fileName:   file1Name,
			keyword:    "is",
			ignoreCase: true,
			result:     GrepResult{MatchedLines: []string{"this", "is", "Is"}},
			// result:     GrepResult{Path: file1Name, MatchedLines: []string{"this", "is", "Is"}},
			expErr:     nil,
		},
		{
			name:             "greps a multi-line file lines with lines before match",
			fileName:         file3Name,
			keyword:          "match",
			ignoreCase:       false,
			linesBeforeMatch: 2,
			result:           GrepResult{MatchedLines: []string{"line4", "line5", "line6 match1"}},
			// result:           GrepResult{Path: file3Name, MatchedLines: []string{"line4", "line5", "line6 match1"}},
			expErr:           nil,
		},
		{
			name:             "greps a multi-line file lines with lines before match",
			fileName:         file4Name,
			keyword:          "match",
			ignoreCase:       false,
			linesBeforeMatch: 2,
			result:           GrepResult{MatchedLines: []string{"line4", "line5", "line6 match1", "line5", "line6 match1", "line7 match2"}},
			// result:           GrepResult{Path: file4Name, MatchedLines: []string{"line4", "line5", "line6 match1", "line5", "line6 match1", "line7 match2"}},
			expErr:           nil,
		},
		{
			name:             "greps a multi-line file lines with lines after match",
			fileName:         file4Name,
			keyword:          "match",
			ignoreCase:       false,
			linesAfterMatch:  1,
			result:           GrepResult{MatchedLines: []string{"line6 match1", "line7 match2", "line7 match2", "line8"}},
			// result:           GrepResult{Path: file4Name, MatchedLines: []string{"line6 match1", "line7 match2", "line7 match2", "line8"}},
			expErr:           nil,
		},
		{
			name:       "greps a multi-line file line with single count",
			fileName:   file3Name,
			keyword:    "match",
			ignoreCase: false,
			lineCount:  true,
			result:     GrepResult{LineCount: 1},
			// result:     GrepResult{Path: file3Name, LineCount: 1},
			// result: GrepResult{MatchedLines: []string{"line6 match1"}, LineCount: 1},
			expErr: nil,
		},
		{
			name:       "greps a multi-line file line with double count",
			fileName:   file4Name,
			keyword:    "match",
			ignoreCase: false,
			lineCount:  true,
			result:     GrepResult{LineCount: 2},
			// result:     GrepResult{Path: file4Name, LineCount: 2},
			// result: GrepResult{MatchedLines: []string{"line6 match1", "line7 match2"}, LineCount: 2},
			expErr: nil,
		},
		{
			name:    "reads from stdin",
			stdin:   []byte("this\nis\na\nfile"),
			keyword: "is",
			result:  GrepResult{MatchedLines: []string{"this", "is"}},
			expErr:  nil,
		},
		{
			name:     "reads a file with permission error",
			fileName: file2Name,
			expErr:   fs.ErrPermission,
		},
		{
			name:     "reads an empty directory",
			fileName: dir1Name,
			expErr:   ErrIsDirectory,
		},
		{
			name:     "reads a non-existent file",
			fileName: "non-existent-file.txt",
			expErr:   fs.ErrNotExist,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := GrepOptions{Path: tc.fileName, Stdin: bytes.NewReader(tc.stdin), Keyword: tc.keyword, IgnoreCase: tc.ignoreCase, LinesBeforeMatch: tc.linesBeforeMatch, LinesAfterMatch: tc.linesAfterMatch, LineCount: tc.lineCount}
			got := Grep(testFS, options)
			want := tc.result

			if tc.expErr != nil {
				if got.Error == nil {
					t.Fatalf("Expected an error but didn't got one")
				}

				if !errors.Is(got.Error, tc.expErr) {
					t.Fatalf("Expected error %q but got %q", tc.expErr.Error(), got.Error.Error())
				}

				return
			}

			if got.Error != nil {
				t.Fatalf("Didn't expected an error: %v", got.Error)
			}

			// checking matched lines
			if !slices.Equal(got.MatchedLines, want.MatchedLines) {
				t.Errorf("Expected %v but got %v", want.MatchedLines, got.MatchedLines)
			}

			// checking matched line count
			if got.LineCount != want.LineCount {
				t.Errorf("Expected line count %d but got %d", want.LineCount, got.LineCount)
			}
		})
	}
}

func TestSearchStringR(t *testing.T) {
	var testFS fstest.MapFS = make(map[string]*fstest.MapFile)
	testFS["testdata"] = &fstest.MapFile{Data: nil, Mode: fs.ModeDir}
	testFS["testdata/test1.txt"] = &fstest.MapFile{Data: []byte("Dummy Line\nthis is a test file\none can test a program by running test cases"), Mode: 0755}
	testFS["testdata/filexyz.txt"] = &fstest.MapFile{Data: []byte("no matches here"), Mode: 0755}
	testFS["testdata/inner/test1.txt"] = &fstest.MapFile{Data: []byte("dummy file"), Mode: 0755}
	testFS["testdata/inner/test2.txt"] = &fstest.MapFile{Data: []byte("this file contains a test line"), Mode: 0755}

	testCases := []struct {
		name       string
		path       string
		keyword    string
		ignoreCase bool
		linesBeforeMatch int
		lineCount bool
		result     []GrepResult
	}{
		{
			name: "greps inside a directory with -r",
			path: "testdata",
			keyword: "test",
			ignoreCase: false,
			result: []GrepResult{
				{Path:"testdata/test1.txt", MatchedLines: []string{"this is a test file", "one can test a program by running test cases"}},
				{Path:"testdata/inner/test2.txt", MatchedLines: []string{"this file contains a test line"}},
			},
			// result: []GrepResult{
			// 	GrepResult{Path:"testdata/test1.txt", MatchedLines: []string{"this is a test file", "one can test a program by running test cases"}, LineCount: 2},
			// 	GrepResult{Path:"testdata/inner/test2.txt", MatchedLines: []string{"this file contains a test line"}, LineCount: 1},
			// },
		},
		{
			name: "greps inside a directory with -r with lines before match option",
			path: "testdata",
			keyword: "test",
			ignoreCase: false,
			linesBeforeMatch: 1,
			result: []GrepResult{
				{Path:"testdata/test1.txt", MatchedLines: []string{"Dummy Line", "this is a test file", "this is a test file", "one can test a program by running test cases"}},
				{Path:"testdata/inner/test2.txt", MatchedLines: []string{"this file contains a test line"}},
			},
			// result: []GrepResult{
			// 	{Path:"testdata/test1.txt", MatchedLines: []string{"Dummy Line", "this is a test file", "this is a test file", "one can test a program by running test cases"}, LineCount: 4},
			// 	{Path:"testdata/inner/test2.txt", MatchedLines: []string{"this file contains a test line"}, LineCount: 1},
			// },
		},
		{
			name: "greps inside a directory with -r with line count option",
			path: "testdata",
			keyword: "test",
			ignoreCase: false,
			lineCount: true,
			result: []GrepResult{
				{Path:"testdata/test1.txt", LineCount: 2},
				{Path:"testdata/inner/test2.txt", LineCount: 1},
			},
			// result: []GrepResult{
			// 	GrepResult{Path:"testdata/test1.txt", MatchedLines: []string{"this is a test file", "one can test a program by running test cases"}, LineCount: 2},
			// 	GrepResult{Path:"testdata/inner/test2.txt", MatchedLines: []string{"this file contains a test line"}, LineCount: 1},
			// },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := GrepOptions{Path: tc.path, Keyword: tc.keyword, IgnoreCase: tc.ignoreCase, LinesBeforeMatch: tc.linesBeforeMatch, LineCount: tc.lineCount}
			got := GrepR(testFS, options)
			want := tc.result

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

			// if !reflect.DeepEqual(got, want) {
			// 	t.Errorf("Expected %v but got %v", want, got)
			// }

			if len(got) != len(want) {
				t.Errorf("Expected length %d but got %d", len(want), len(got))
			}
		})
	}
}
