package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
	"os"

	grep "github.com/one2n-go-bootcamp/grep/pkg"
)

func run(fSys fs.FS, stdin io.Reader, args []string, fileName string, linesBeforeMatch int, ignoreCase, searchDir, lineCount bool) string {
	var result [][]string
	var err error
	var output string

	options := grep.GrepOptions{
		Stdin: stdin,
		Keyword: args[0],
		IgnoreCase: ignoreCase,
		LinesBeforeMatch: linesBeforeMatch,
		LineCount: lineCount,
	}

	if linesBeforeMatch > 0 {
		options.LinesBeforeMatch = linesBeforeMatch
	}

	if len(args) > 1 {
		result, err = grep.GrepR(fSys, options)
	} else {
		options.Path = args[1]

		if grepResult, err := grep.Grep(fSys, options); err != nil {
			return err.Error()
		} else {
			result = append(result, grepResult)
		}
	}

	if err != nil {
		return err.Error()
	}

	// output := strings.Join(result, "\n")
	for _, r := range result {
		out := strings.Join(r, ",")
		output += out
	}


	if fileName != "" {
		err := writeToFile(fileName, output)
		if err != nil {
			return err.Error()
		}
		return ""
	}

	return output
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
