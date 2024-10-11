package grep

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"
	"testing/iotest"
)

var testFS = fstest.MapFS{
	"file1.txt": {Data: []byte(""), Mode: 0755},
	"file2.txt": {Data: []byte("single_line"), Mode: 0755},
	"file3.txt": {Data: []byte("single line\nand\ndouble line\nin\nfile"), Mode: 0755},
	"file4.txt": {Data: []byte("\nI love mangoes,\tapples- but it applies to most fruits.\n??--ww"), Mode: 0755},
	"file5.txt": {Data: []byte("this file got permisson error"), Mode: 0000},
	"file6.txt": {Data: []byte("this\nis\na\nfile\nIs"), Mode: 0755},
	"file7.txt": {Data: []byte("line1\nline2\nline3\nline4\nline5\nline6 match1\nline7\nline8\nline9"), Mode: 0755},
	"file8.txt": {Data: []byte("line1\nline2\nline3\nline4\nline5\nline6 match1\nline7 match2\nline8\nline9\nline10"), Mode: 0755},
	"dir1":      {Mode: fs.ModeDir},
	"testdata":  {Data: nil, Mode: fs.ModeDir},
	"testdata/test1.txt": {
		Data: []byte("Dummy Line\nthis is a test file\none can test a program by running test cases"),
		Mode: 0755,
	},
	"testdata/filexyz.txt":     {Data: []byte("no matches here"), Mode: 0755},
	"testdata/inner/test1.txt": {Data: []byte("dummy file"), Mode: 0755},
	"testdata/inner/test2.txt": {Data: []byte("this file contains a test line"), Mode: 0755},
	"testdata/mdFile.md":       {Data: []byte("this is a test md"), Mode: 0755},
	"testdata/logFile.log":     {Data: []byte("this is a test log"), Mode: 0755},
}

func TestGrep(t *testing.T) {
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
			fileName:   "file6.txt",
			keyword:    "is",
			ignoreCase: false,
			result:     GrepResult{MatchedLines: []string{"this", "is"}},
			expErr:     nil,
		},
		{
			name:       "greps a multi-line file text sensitive",
			fileName:   "file6.txt",
			keyword:    "is",
			ignoreCase: true,
			result:     GrepResult{MatchedLines: []string{"this", "is", "Is"}},
			expErr:     nil,
		},
		{
			name:             "greps a multi-line file lines with lines before match",
			fileName:         "file7.txt",
			keyword:          "match",
			ignoreCase:       false,
			linesBeforeMatch: 2,
			result:           GrepResult{MatchedLines: []string{"line4", "line5", "line6 match1"}},
			expErr:           nil,
		},
		{
			name:             "greps a multi-line file lines with lines before match",
			fileName:         "file8.txt",
			keyword:          "match",
			ignoreCase:       false,
			linesBeforeMatch: 2,
			result:           GrepResult{MatchedLines: []string{"line4", "line5", "line6 match1", "line5", "line6 match1", "line7 match2"}},
			expErr:           nil,
		},
		{
			name:            "greps a multi-line file lines with lines after match",
			fileName:        "file8.txt",
			keyword:         "match",
			ignoreCase:      false,
			linesAfterMatch: 1,
			result:          GrepResult{MatchedLines: []string{"line6 match1", "line7 match2", "line7 match2", "line8"}},
			expErr:          nil,
		},
		{
			name:       "greps a multi-line file line with single count",
			fileName:   "file7.txt",
			keyword:    "match",
			ignoreCase: false,
			lineCount:  true,
			result:     GrepResult{LineCount: 1},
			expErr:     nil,
		},
		{
			name:       "greps a multi-line file line with double count",
			fileName:   "file8.txt",
			keyword:    "match",
			ignoreCase: false,
			lineCount:  true,
			result:     GrepResult{LineCount: 2},
			expErr:     nil,
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
			fileName: "file5.txt",
			expErr:   fs.ErrPermission,
		},
		{
			name:     "reads an empty directory",
			fileName: "dir1",
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
			options := GrepOption{Path: tc.fileName, Stdin: bytes.NewReader(tc.stdin), Keyword: tc.keyword, IgnoreCase: tc.ignoreCase, LinesBeforeMatch: tc.linesBeforeMatch, LinesAfterMatch: tc.linesAfterMatch, LineCount: tc.lineCount}
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

func TestGrepR(t *testing.T) {
	testCases := []struct {
		name             string
		path             string
		keyword          string
		ignoreCase       bool
		linesBeforeMatch int
		lineCount        bool
		includeExt       []string
		excludeExt       []string
		result           []GrepResult
	}{
		{
			name:       "greps inside a directory",
			path:       "testdata",
			keyword:    "test",
			ignoreCase: false,
			result: []GrepResult{
				{
					Path:         "testdata/test1.txt",
					MatchedLines: []string{"this is a test file", "one can test a program by running test cases"},
				},
				{
					Path:         "testdata/inner/test2.txt",
					MatchedLines: []string{"this file contains a test line"},
				},
				{
					Path:         "testdata/mdFile.md",
					MatchedLines: []string{"this is a test md"},
				},
				{
					Path:         "testdata/logFile.log",
					MatchedLines: []string{"this is a test log"},
				},
			},
		},
		{
			name:             "greps inside a directory with lines before match and include txt extension option",
			path:             "testdata",
			keyword:          "test",
			ignoreCase:       false,
			linesBeforeMatch: 1,
			includeExt:       []string{"txt"},
			result: []GrepResult{
				{
					Path: "testdata/test1.txt",
					MatchedLines: []string{
						"Dummy Line", "this is a test file",
						"this is a test file",
						"one can test a program by running test cases",
					},
				},
				{
					Path:         "testdata/inner/test2.txt",
					MatchedLines: []string{"this file contains a test line"},
				},
			},
		},
		{
			name:       "greps inside a directory with line count and include txt extension option",
			path:       "testdata",
			keyword:    "test",
			ignoreCase: false,
			lineCount:  true,
			includeExt: []string{"txt"},
			result: []GrepResult{
				{
					Path:      "testdata/test1.txt",
					LineCount: 2,
				},
				{
					Path:      "testdata/inner/test2.txt",
					LineCount: 1,
				},
			},
		},
		{
			name:       "greps inside a directory with exclude extension option",
			path:       "testdata",
			keyword:    "test",
			ignoreCase: false,
			lineCount:  true,
			excludeExt: []string{"md", "log"},
			result: []GrepResult{
				{
					Path:      "testdata/test1.txt",
					LineCount: 2,
				},
				{
					Path:      "testdata/inner/test2.txt",
					LineCount: 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := GrepOption{Path: tc.path, Keyword: tc.keyword, IgnoreCase: tc.ignoreCase, LinesBeforeMatch: tc.linesBeforeMatch, LineCount: tc.lineCount, ExcludeExt: tc.excludeExt, IncludeExt: tc.includeExt}
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

func TestGetReader(t *testing.T) {
	tt := []struct {
		name            string
		option          GrepOption
		expTextInReader string
		expErr          error
	}{
		{
			name: "nomal happy case with valid path",
			option: GrepOption{
				Path:  "file2.txt",
				Stdin: nil,
			},
			expTextInReader: "single_line",
			expErr:          nil,
		},
		{
			name: "nomal happy case with empty path and valid stdin",
			option: GrepOption{
				Path:  "",
				Stdin: bytes.NewReader([]byte("text from stdin\n here")),
			},
			expTextInReader: "text from stdin\n here",
			expErr:          nil,
		},
		{
			name: "error unhappy case with invalid path of file",
			option: GrepOption{
				Path:  "invalid_file.txt",
				Stdin: nil,
			},
			expTextInReader: "",
			expErr:          fs.ErrNotExist,
		},
		{
			name: "error unhappy case with valid path of directory",
			option: GrepOption{
				Path:  "dir1",
				Stdin: nil,
			},
			expTextInReader: "",
			expErr:          ErrIsDirectory,
		},
		{
			name: "error unhappy case with valid path of file but permisson error",
			option: GrepOption{
				Path:  "file5.txt",
				Stdin: nil,
			},
			expTextInReader: "",
			expErr:          fs.ErrPermission,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			reader, cleanup, err := getReader(testFS, tc.option)
			if tc.expErr != nil {
				if err == nil {
					t.Fatalf("Expected error %q but got nil", tc.expErr.Error())
				}

				if !errors.Is(err, tc.expErr) {
					t.Fatalf("Expected error %q but got %q", tc.expErr.Error(), err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %q", err.Error())
			}

			// checking if the reader has the same text as we expected
			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("Unexpected error while reading from io.Reader, %q", err.Error())
			}
			if string(data) != tc.expTextInReader {
				t.Errorf("Expected the reader to have %q but got %q", tc.expTextInReader, string(data))
			}

			// TODO: testing the behaviour of cleanup function
			// maybe we have to find a way to check if the file is still open (that would be too coupled to the implementation)
			// we can also do some kind of mock counter thing where the cleanup function would be injected from arguments and we will check if the counter has been incremented
			// very unclear what to do here
			cleanup()
		})
	}
}

func TestIsValid(t *testing.T) {
	testFS := fstest.MapFS{
		"file1.txt": {Data: []byte(""), Mode: 0755},
		"file2.txt": {Data: []byte("single_line"), Mode: 0755},
		"file3.txt": {Data: []byte("single line\nand\ndouble line\nin\nfile"), Mode: 0755},
		"file4.txt": {Data: []byte("\nI love mangoes,\tapples- but it applies to most fruits.\n??--ww"), Mode: 0755},
		"file5.txt": {Data: []byte("this file got permisson error"), Mode: 0000},
		"file6.txt": {Data: []byte("dummy file 6"), Mode: 0755},
		"dir1":      {Mode: fs.ModeDir},
	}

	tt := []struct {
		name            string
		option          GrepOption
		expOut          bool
		expTextInReader string
		expErr          error
	}{
		{
			name:   "nomal happy case with valid path",
			expOut: true,
			option: GrepOption{
				Path: "file2.txt",
			},
			expErr: nil,
		},
		{
			name:   "error unhappy case with empty path",
			expOut: false,
			option: GrepOption{
				Path: "",
			},
			expErr: fs.ErrNotExist,
		},
		{
			name: "error unhappy case with invalid path of file",
			option: GrepOption{
				Path: "xyz.txt",
			},
			expOut: false,
			expErr: fs.ErrNotExist,
		},
		{
			name: "error unhappy case with valid path of directory",
			option: GrepOption{
				Path: "dir1",
			},
			expOut: false,
			expErr: ErrIsDirectory,
		},
		{
			name: "error unhappy case with valid path of file but permisson error",
			option: GrepOption{
				Path: "file5.txt",
			},
			expOut: false,
			expErr: fs.ErrPermission,
		},
		{
			name: "exclude file extension",
			option: GrepOption{
				Path:       "file6.txt",
				ExcludeExt: []string{"txt"},
			},
			expOut: false,
			expErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := isValid(testFS, tc.option)

			if tc.expErr != nil {
				if gotErr == nil {
					t.Fatalf("Expected error %q but got nil", tc.expErr.Error())
				}

				if !errors.Is(gotErr, tc.expErr) {
					t.Fatalf("Expected error %q but got %q", tc.expErr.Error(), gotErr.Error())
				}

				return
			}

			if gotErr != nil {
				t.Fatalf("Unexpected error: %q", gotErr.Error())
			}

			if got != tc.expOut {
				t.Errorf("Expected %v but got %v", tc.expOut, got)
			}
		})
	}
}

func TestNormalisePathFromRoot(t *testing.T) {
	tt := []struct {
		name     string
		rootPath string
		dirPath  string
		expOut   string
	}{
		{
			name:     "happy path - 1",
			rootPath: "go-grep/testdata/cmd_test/inner/test2.txt",
			dirPath:  "../testdata/cmd_test",
			expOut:   "../testdata/cmd_test/inner/test2.txt",
		},
		{
			name:     "happy path - 2",
			rootPath: "go-grep/testdata/cmd_test/inner/testdata/cmd_test/test2.txt",
			dirPath:  "../testdata/cmd_test",
			expOut:   "../testdata/cmd_test/inner/testdata/cmd_test/test2.txt",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := normalisePathFromRoot(tc.rootPath, tc.dirPath)

			if got != tc.expOut {
				t.Errorf("Expected %v but got %v", tc.expOut, got)
			}
		})
	}
}

func TestSearchString(t *testing.T) {
	tt := []struct {
		name   string
		reader io.Reader
		option GrepOption
		expOut []string
		expErr error
	}{
		{
			name:   "normal search",
			reader: bytes.NewReader([]byte("Dummy Line\nthis is a test file\none can test a program by running test cases")),
			option: GrepOption{
				Keyword: "test",
			},
			expOut: []string{
				"this is a test file",
				"one can test a program by running test cases",
			},
			expErr: nil,
		},
		{
			name:   "search with lines before match option",
			reader: bytes.NewReader([]byte("Dummy Line\nthis is a test file\none can test a program by running test cases")),
			option: GrepOption{
				Keyword:          "test",
				LinesBeforeMatch: 1,
			},
			expOut: []string{
				"Dummy Line",
				"this is a test file",
				"this is a test file",
				"one can test a program by running test cases",
			},
			expErr: nil,
		},
		{
			name:   "search with lines before match and lines after match option",
			reader: bytes.NewReader([]byte("Dummy Line\nthis is a test file\none can test a program by running test cases")),
			option: GrepOption{
				Keyword:          "test",
				LinesBeforeMatch: 1,
				LinesAfterMatch:  1,
			},
			expOut: []string{
				"Dummy Line",
				"this is a test file",
				"one can test a program by running test cases",
				"this is a test file",
				"one can test a program by running test cases",
			},
			expErr: nil,
		},
		{
			name:   "search with lines before match and lines after match option",
			reader: bytes.NewReader([]byte("Dummy Line\nthis is a test file\none can test a program by running test cases")),
			option: GrepOption{
				Keyword:          "test",
				LinesBeforeMatch: 1,
				LinesAfterMatch:  1,
			},
			expOut: []string{
				"Dummy Line",
				"this is a test file",
				"one can test a program by running test cases",
				"this is a test file",
				"one can test a program by running test cases",
			},
			expErr: nil,
		},
		{
			name:   "loaded error prone reader",
			reader: iotest.ErrReader(io.ErrUnexpectedEOF),
			option: GrepOption{
				Keyword: "test",
			},
			expOut: []string{
				"this is a test file",
				"one can test a program by running test cases",
			},
			expErr: io.ErrUnexpectedEOF,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := searchString(tc.reader, tc.option)
			if tc.expErr != nil {
				if gotErr == nil {
					t.Errorf("Expected error %v but got nil", tc.expErr)
				}

				if !errors.Is(gotErr, tc.expErr) {
					t.Errorf("Expected error %v but got %v", tc.expErr, gotErr)
				}

				return
			}

			if gotErr != nil {
				t.Errorf("Unexpected error: %v", gotErr)
			}

			if !slices.Equal(tc.expOut, got) {
				t.Errorf("Expected %v but got %v", tc.expOut, got)
			}
		})
	}
}
