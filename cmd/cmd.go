package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	grep "github.com/one2n-go-bootcamp/go-grep/pkg"
)

// to handle the grep input from user
type GrepInput struct {
	keyword          string
	path             string
	fileWriteName    string
	linesBeforeMatch int
	linesAfterMatch  int
	ignoreCase       bool
	searchDir        bool
	lineCount        bool
	stdin            io.Reader
	output           io.Writer
}

func run(fSys fs.FS, input *GrepInput) {
	option := grep.GrepOption{}

	// stdin case
	if input.path == "" {
		option.Stdin = input.stdin
		option.Keyword = input.keyword
		option.FileWName = input.fileWriteName
		option.IgnoreCase = input.ignoreCase
		option.LinesBeforeMatch = input.linesBeforeMatch
		option.LinesAfterMatch = input.linesAfterMatch
		option.SearchDir = input.searchDir
		option.LineCount = input.lineCount
	} else {
		// file case
		fullPath, err := getFullPath(fSys, input.path)
		if err != nil {
			fmt.Println(err)
			return
		}

		option.Keyword = input.keyword
		option.OrigPath = input.path
		option.Path = fullPath
		option.FileWName = input.fileWriteName
		option.IgnoreCase = input.ignoreCase
		option.LinesBeforeMatch = input.linesBeforeMatch
		option.LinesAfterMatch = input.linesAfterMatch
		option.SearchDir = input.searchDir
		option.LineCount = input.lineCount
	}

	var result []grep.GrepResult
	if input.searchDir {
		result = grep.GrepR(fSys, option)
	} else {
		grepResult := grep.Grep(fSys, option)
		if grepResult.Error != nil {
			fmt.Fprintln(input.output, grepResult.Error.Error())
			return
		}
		result = append(result, grepResult)
	}

	var outputArr []string
	// preparing to print the result on the basis of options
	for _, res := range result {
		if input.searchDir && option.LineCount {
			outputArr = append(outputArr, fmt.Sprintf("%s:%d\n", res.Path, res.LineCount))
		} else if input.searchDir && !option.LineCount {
			for _, line := range res.MatchedLines {
				outputArr = append(outputArr, fmt.Sprintf("%s:%s\n", res.Path, line))
			}
		} else {
			for _, line := range res.MatchedLines {
				outputArr = append(outputArr, fmt.Sprintf("%s\n", line))
			}
		}
	}

	// writing to file if file name was passed
	if input.fileWriteName != "" {
		err := writeToFile(input.fileWriteName, strings.Join(outputArr, ""))
		if err != nil {
			fmt.Fprint(input.output, err.Error())
			return
		}
		return
	}

	fmt.Fprint(input.output, strings.Join(outputArr, ""))
}

func writeToFile(filePath string, content string) error {
	// check if file exists
	_, err := os.Stat(filePath)
	if err == nil {
		return fmt.Errorf("%s: %w", filePath, os.ErrExist)
	}

	// create file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// write to file
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

// gets the path from fSys (/ in this case) to the arg
func getFullPath(fSys fs.FS, arg string) (relPath string, err error) {
	absPath, err := filepath.Abs(filepath.Clean(arg))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	root := fmt.Sprintf("%s", fSys)
	relPath, err = filepath.Rel(root, absPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}
