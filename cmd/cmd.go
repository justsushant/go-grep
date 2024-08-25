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

func run(fSys fs.FS, stdin io.Reader, out io.Writer, keyword, path, fileWName string, linesBeforeMatch int, linesAfterMatch int, ignoreCase, searchDir, lineCount bool) {
	option := grep.GrepOptions{}

	// stdin case
	if path == "" {
		option.Stdin = stdin
		option.Keyword = keyword
		option.FileWName = fileWName
		option.IgnoreCase = ignoreCase
		option.LinesBeforeMatch = linesBeforeMatch
		option.LinesAfterMatch = linesAfterMatch
		option.SearchDir = searchDir
		option.LineCount = lineCount
	} else {
		// file case
		fullPath, err := getFullPath(fSys, path)
		if err != nil {
			fmt.Println(err)
			return
		}

		option.Keyword = keyword
		option.OrigPath = path
		option.Path = fullPath
		option.FileWName = fileWName
		option.IgnoreCase = ignoreCase
		option.LinesBeforeMatch = linesBeforeMatch
		option.LinesAfterMatch = linesAfterMatch
		option.SearchDir = searchDir
		option.LineCount = lineCount
	}

	var result []grep.GrepResult
	if searchDir {
		result = grep.GrepR(fSys, option)
	} else {
		grepResult := grep.Grep(fSys, option)
		if grepResult.Error != nil {
			fmt.Fprint(out, grepResult.Error.Error())
			return
		}
		result = append(result, grepResult)
	}

	var outputArr []string
	// preparing to print the result on the basis of options
	for _, res := range result {
		if searchDir && option.LineCount {
			outputArr = append(outputArr, fmt.Sprintf("%s:%d\n", res.Path, res.LineCount))
		} else if searchDir && !option.LineCount {
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
	if fileWName != "" {
		err := writeToFile(fileWName, strings.Join(outputArr, ""))
		if err != nil {
			fmt.Fprint(out, err.Error())
			return
		}
		return
	}

	fmt.Fprint(out, strings.Join(outputArr, ""))
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
