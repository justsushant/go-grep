package cmd

import (
	"bytes"
	// "errors"
	"io"

	"io/fs"
	"os"
	"strings"
	"testing"

	grep "github.com/one2n-go-bootcamp/go-grep/pkg"
)

// since run function integrates all the other functions, so actual files are used
func TestRun(t *testing.T) {
	testCases := []struct {
		name             string
		stdin            io.Reader
		fileWName        string
		path             string
		keyword          string
		ignoreCase       bool
		linesBeforeMatch int
		linesAfterMatch int
		searchDir        bool
		lineCount        bool
		result           [][]string
		expErr           error
	}{
		{
			name:    "greps on a multi-line file",
			path:    "../testdata/cmd_test/test2.txt",
			keyword: "match",
			result:  [][]string{{"no matches here"}},
		},
		{
			name:    "greps on a multi-line file without matches",
			path:    "../testdata/cmd_test/test2.txt",
			keyword: "vibgyor",
			result:  [][]string{},
		},
		{
			name:    "greps on stdin",
			stdin:   bytes.NewReader([]byte("you will find\nno matches here\nwhatsoever")),
			keyword: "match",
			result:  [][]string{{"no matches here"}},
		},
		{
			name:    "greps on a non-existent file",
			path:    "../testdata/cmd_test/non-existent-file.txt",
			keyword: "vibgyor",
			expErr:  fs.ErrNotExist,
		},
		{
			name:    "greps on a directory",
			path:    "../testdata/cmd_test/inner",
			keyword: "vibgyor",
			expErr:  grep.ErrIsDirectory,
		},
		// {
		// 	name: "reads a file with permission error",
		// 	path: "testdata/cmd_test/perm_err/test1.txt",
		// 	expErr: fs.ErrPermission,
		// },
		{
			name:      "greps inside a directory with -r",
			path:      "../testdata/cmd_test",
			keyword:   "test",
			searchDir: true,
			result: [][]string{
				{"../testdata/cmd_test/test1.txt:this is a test file", "../testdata/cmd_test/test1.txt:one can test a program by running test cases"},
				{"../testdata/cmd_test/inner/test2.txt:this file contains a test line"},
			},
		},
		{
			name:      "greps inside a directory with -r without matches",
			path:      "../testdata/cmd_test",
			keyword:   "vibgyor",
			searchDir: true,
			result:    [][]string{},
		},
		{
			name:             "greps inside a directory with -r with 1 line before match option",
			path:             "../testdata/cmd_test",
			keyword:          "test",
			searchDir:        true,
			linesBeforeMatch: 1,
			result: [][]string{
				{"../testdata/cmd_test/test1.txt:Dummy Line", "../testdata/cmd_test/test1.txt:this is a test file", "../testdata/cmd_test/test1.txt:this is a test file", "../testdata/cmd_test/test1.txt:one can test a program by running test cases"},
				{"../testdata/cmd_test/inner/test2.txt:this file contains a test line"},
			},
		},
		{
			name:             "greps inside a directory with -r with 1 line after match option",
			path:             "../testdata/cmd_test",
			keyword:          "test",
			searchDir:        true,
			linesAfterMatch: 1,
			result: [][]string{
				{
					"../testdata/cmd_test/test1.txt:this is a test file",
					"../testdata/cmd_test/test1.txt:this is a test file",
					"../testdata/cmd_test/test1.txt:one can test a program by running test cases",
					"../testdata/cmd_test/test1.txt:something here",
				},
				{
					"../testdata/cmd_test/inner/test2.txt:this file contains a test line",
					"../testdata/cmd_test/inner/test2.txt:nothing here",
				},
			},
		},
		{
			name:      "greps inside a directory with -r with line count option",
			path:      "../testdata/cmd_test",
			keyword:   "test",
			searchDir: true,
			lineCount: true,
			result:    [][]string{{"../testdata/cmd_test/test1.txt:2"}, {"../testdata/cmd_test/inner/test2.txt:1"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := os.DirFS("/")
			var got bytes.Buffer
			want := getExpectedOutput(t, tc.result)

			run(fs, tc.stdin, &got, tc.keyword, tc.path, tc.fileWName, tc.linesBeforeMatch, tc.linesAfterMatch, tc.ignoreCase, tc.searchDir, tc.lineCount)

			// checking for error
			if tc.expErr != nil {
				if !strings.Contains(got.String(), tc.expErr.Error()) {
					t.Fatalf("Expected error %q not found in the final output %q\n", tc.expErr.Error(), got.String())
				}
				return
			}

			// length check of both expected and resultant
			// trimming to remove the last new line character
			if len(strings.Split(strings.TrimSpace(got.String()), "\n")) != len(strings.Split(want, "\n")) {
				t.Errorf("Expected number of line %d but got %d", len(strings.Split(want, "\n")), len(strings.Split(got.String(), "\n")))
			}

			// checking if each line wanted is present in output
			for _, w := range strings.Split(want, "\n") {
				matchFlag := false
				for _, g := range strings.Split(got.String(), "\n") {
					if g == w {
						matchFlag = true
						break
					}
				}

				if !matchFlag {
					t.Errorf("Expected string %q was not found in final output %q", w, got.String())
				}
			}
		})
	}
}

// func TestWriteToFile(t *testing.T) {
// 	testCases := []struct {
// 		name     string
// 		filePath string
// 		content  string
// 		expErr   error
// 	}{
// 		{name: "write to file", filePath: "test.txt", content: "test only", expErr: nil},
// 		{name: "write to already created file", filePath: "test.txt", content: "test only", expErr: os.ErrExist},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			if tc.expErr != nil {
// 				os.Create(tc.filePath)
// 			}

// 			err := writeToFile(tc.filePath, tc.content)
// 			defer os.Remove(tc.filePath)

// 			if tc.expErr != nil {
// 				if err == nil {
// 					t.Fatalf("Expected error but didn't got one")
// 				}

// 				if !errors.Is(err, tc.expErr) {
// 					t.Errorf("Expected error %v but got %v", tc.expErr, err)
// 				}

// 				return
// 			}

// 			if err != nil {
// 				t.Fatalf("Unexpected error: %v", err)
// 			}

// 			data, err := os.ReadFile(tc.filePath)
// 			if err != nil {
// 				t.Fatalf("Unexpected error: %v", err)
// 			}

// 			if string(data) != tc.content {
// 				t.Errorf("Expected %q but got %q", string(data), tc.content)
// 			}
// 		})
// 	}
// }

func getExpectedOutput(t *testing.T, result [][]string) string {
	t.Helper()
	var wantArr []string
	for _, res := range result {
		wantArr = append(wantArr, strings.Join(res, "\n"))
	}
	return strings.Join(wantArr, "\n")
}
