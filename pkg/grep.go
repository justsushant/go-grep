package grep

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

var (
	ErrIsDirectory = errors.New("is a directory")
)

func SearchString(fSys fs.FS, path string, stdin io.Reader, keyword string, ignoreCase bool, searchDir bool) ([]string, error) {
	err := isValid(fSys, path, searchDir)
	if err != nil {
		return nil, err
	}

	scanner, err, cleanup := getScanner(fSys, path, stdin)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	

	result, err := search(scanner, keyword, ignoreCase)
	if err != nil {
		return nil, err
	}

	return result, nil	
}

func isValid(fSys fs.FS, path string, searchDir bool) error {
	if path == "" {
		return nil
	}

	fileInfo, err := fs.Stat(fSys, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", path, fs.ErrNotExist)
		}
		return fmt.Errorf("%s: %w", path, err)
	}

	if !searchDir {
		// checks for directory
		if fileInfo.IsDir() {
			return fmt.Errorf("%s: %w", path, ErrIsDirectory)
		}

		// checks for permissions
		// looks hacky, might have to change later
		if fileInfo.Mode().Perm()&400 == 0 {
			return fmt.Errorf("%s: %w", path, fs.ErrPermission)
		}
	}

	return nil
}

func getScanner(fSys fs.FS, path string, stdin io.Reader) (*bufio.Scanner, error, func()) {
	if path != "" {
		// open file
		file, err := fSys.Open(path)
		if err != nil {
			return nil, err, func(){}
		}

		return bufio.NewScanner(file), nil, func() {file.Close()}
	}

	return bufio.NewScanner(stdin), nil, func(){}
}

func search(scanner *bufio.Scanner, keyword string, ignoreCase bool) ([]string, error) {
	var result []string
	if ignoreCase {
		keyword = strings.ToLower(keyword)
	}

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var line string
		line = scanner.Text()

		if ignoreCase {
			line = strings.ToLower(scanner.Text())
		}
		
		if strings.Contains(line, keyword) {
			result = append(result, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}