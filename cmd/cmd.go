package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
	"os"

	grep "github.com/one2n-go-bootcamp/grep/pkg"
)

func run(fSys fs.FS, stdin io.Reader, args []string, fileName string, isCaseSensitive bool) string {
	var result []string
	var err error

	if len(args) > 1 {
		result, err = grep.SearchString(fSys, args[1], stdin, args[0], isCaseSensitive)
	} else {
		result, err = grep.SearchString(fSys, "", stdin, args[0], isCaseSensitive)
	}

	if err != nil {
		return err.Error()
	}

	output := strings.Join(result, "\n")
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
