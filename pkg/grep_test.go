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
	// setting the test file system
	testFS := fstest.MapFS{}
	testFS["file1.txt"] = &fstest.MapFile{
		Data: []byte("this\nis\na\nfile\nIs"), 
		Mode: 0755,
	}
	testFS["file2.txt"] = &fstest.MapFile{
		Data: []byte{}, 
		Mode: 0000,
	}
	testFS["file3.txt"] = &fstest.MapFile{
		Data: []byte("line1\nline2\nline3\nline4\nline5\nline6 match1\nline7\nline8\nline9"), 
		Mode: 0755,
	}
	testFS["file4.txt"] = &fstest.MapFile{
		Data: []byte("line1\nline2\nline3\nline4\nline5\nline6 match1\nline7 match2\nline8\nline9\nline10"), 
		Mode: 0755,
	}
	testFS["testDir"] = &fstest.MapFile{Data: nil, Mode: fs.ModeDir}

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
			fileName:   "file1.txt",
			keyword:    "is",
			ignoreCase: false,
			result:     GrepResult{MatchedLines: []string{"this", "is"}},
			expErr:     nil,
		},
		{
			name:       "greps a multi-line file text sensitive",
			fileName:   "file1.txt",
			keyword:    "is",
			ignoreCase: true,
			result:     GrepResult{MatchedLines: []string{"this", "is", "Is"}},
			expErr:     nil,
		},
		{
			name:             "greps a multi-line file lines with lines before match",
			fileName:         "file3.txt",
			keyword:          "match",
			ignoreCase:       false,
			linesBeforeMatch: 2,
			result:           GrepResult{MatchedLines: []string{"line4", "line5", "line6 match1"}},
			expErr:           nil,
		},
		{
			name:             "greps a multi-line file lines with lines before match",
			fileName:         "file4.txt",
			keyword:          "match",
			ignoreCase:       false,
			linesBeforeMatch: 2,
			result:           GrepResult{MatchedLines: []string{"line4", "line5", "line6 match1", "line5", "line6 match1", "line7 match2"}},
			expErr:           nil,
		},
		{
			name:             "greps a multi-line file lines with lines after match",
			fileName:         "file4.txt",
			keyword:          "match",
			ignoreCase:       false,
			linesAfterMatch:  1,
			result:           GrepResult{MatchedLines: []string{"line6 match1", "line7 match2", "line7 match2", "line8"}},
			expErr:           nil,
		},
		{
			name:       "greps a multi-line file line with single count",
			fileName:   "file3.txt",
			keyword:    "match",
			ignoreCase: false,
			lineCount:  true,
			result:     GrepResult{LineCount: 1},
			expErr: nil,
		},
		{
			name:       "greps a multi-line file line with double count",
			fileName:   "file4.txt",
			keyword:    "match",
			ignoreCase: false,
			lineCount:  true,
			result:     GrepResult{LineCount: 2},
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
			fileName: "file2.txt",
			expErr:   fs.ErrPermission,
		},
		{
			name:     "reads an empty directory",
			fileName: "testDir",
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
	testFS["testdata/test1.txt"] = &fstest.MapFile{
		Data: []byte("Dummy Line\nthis is a test file\none can test a program by running test cases"), 
		Mode: 0755,
	}
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
				{
					Path:"testdata/test1.txt",
					MatchedLines: []string{"this is a test file", "one can test a program by running test cases"},
				},
				{
					Path:"testdata/inner/test2.txt",
					MatchedLines: []string{"this file contains a test line"},
				},
			},
		},
		{
			name: "greps inside a directory with -r with lines before match option",
			path: "testdata",
			keyword: "test",
			ignoreCase: false,
			linesBeforeMatch: 1,
			result: []GrepResult{
				{
					Path:"testdata/test1.txt",
					MatchedLines: []string{
						"Dummy Line", "this is a test file",
						"this is a test file",
						"one can test a program by running test cases",
					},
				},
				{
					Path:"testdata/inner/test2.txt",
					MatchedLines: []string{"this file contains a test line"},
				},
			},
		},
		{
			name: "greps inside a directory with -r with line count option",
			path: "testdata",
			keyword: "test",
			ignoreCase: false,
			lineCount: true,
			result: []GrepResult{
				{
					Path:"testdata/test1.txt",
					LineCount: 2,
				},
				{
					Path:"testdata/inner/test2.txt", 
					LineCount: 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := GrepOptions{Path: tc.path, Keyword: tc.keyword, IgnoreCase: tc.ignoreCase, LinesBeforeMatch: tc.linesBeforeMatch, LineCount: tc.lineCount}
			got := GrepR(testFS, options)
			want := tc.result

			// checks for equality by checking each slice since order is not guaranteed
			for _, g := range got {
				matchFlag := false
				for _, w := range want {
					if g.Path == w.Path && slices.Equal(g.MatchedLines, w.MatchedLines) && g.LineCount == w.LineCount {
                        matchFlag = true
                        break
                    }
				}

				if !matchFlag {
					t.Errorf("Expected %v but got %v", want, got)
				}
			}

			if len(got) != len(want) {
				t.Errorf("Expected length %d but got %d", len(want), len(got))
			}
		})
	}
}
