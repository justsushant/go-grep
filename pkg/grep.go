package grep

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	// "io"
	"io/fs"
	"strings"
)

var (
	ErrIsDirectory = errors.New("is a directory")
)

func SearchString(fSys fs.FS, path string, stdin io.Reader, keyword string) ([]string, error) {
	var scanner *bufio.Scanner

	if path != "" {
		err := isValid(fSys, path)
		if err != nil {
			return nil, err
		}

		file, err := fSys.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	} else {
		scanner = bufio.NewScanner(stdin)
	}

	
	var result []string
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, keyword) {
			result = append(result, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil	
}

func isValid(fSys fs.FS, path string) error {
	fileInfo, err := fs.Stat(fSys, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", path, fs.ErrNotExist)
		}
		return fmt.Errorf("%s: %w", path, err)
	}

	// checks for directory
	if fileInfo.IsDir() {
		return fmt.Errorf("%s: %w", path, ErrIsDirectory)
	}

	// checks for permissions
	// looks hacky, might have to change later
	if fileInfo.Mode().Perm()&400 == 0 {
		return fmt.Errorf("%s: %w", path, fs.ErrPermission)
	}

	return nil
}