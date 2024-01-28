package grep

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"sync"
)

var (
	ErrIsDirectory = errors.New("is a directory")
)

type GrepOptions struct {
	path string
	stdin io.Reader
	keyword string
	ignoreCase bool
	linesBeforeMatch int
}

func Grep(fSys fs.FS, options GrepOptions) ([]string, error) {
	r, cleanup, err := getReader(fSys, options.path, options.stdin)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	result, err := searchString(r, options.keyword, options.ignoreCase, options.linesBeforeMatch)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getReader(fSys fs.FS, fileName string, stdin io.Reader) (io.Reader, func(), error) {
	if fileName != "" {
		err := isValid(fSys, fileName)
		if err != nil {
			return nil, nil, err
		}

		file, err := fSys.Open(fileName)
		if err != nil {
			return nil, nil, err
		}
		return file, func() {file.Close()}, nil
	}

	return stdin, func() {}, nil
}

func searchString(r io.Reader,  keyword string, ignoreCase bool, linesBeforeMatch int) ([]string, error) {
	grepBuff := NewGrepBuffer(linesBeforeMatch)
	var result []string
	scanner := bufio.NewScanner(r)

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
			if linesBeforeMatch > 0 {
				result = append(result, grepBuff.Dump()...)
			} 

			result = append(result, scanner.Text())
			
		}
		
		if linesBeforeMatch > 0 {
			grepBuff.Push(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func isValid(fSys fs.FS, fileName string) error {
	fileInfo, err := fs.Stat(fSys, fileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", fileName, fs.ErrNotExist)
		}
		return fmt.Errorf("%s: %w", fileName, err)
	}

	// checks for directory
	if fileInfo.IsDir() {
		return fmt.Errorf("%s: %w", fileName, ErrIsDirectory)
	}

	// checks for permissions
	// looks hacky, might have to change later
	if fileInfo.Mode().Perm()&400 == 0 {
		return fmt.Errorf("%s: %w", fileName, fs.ErrPermission)
	}

	return nil
}

func GrepR(fSys fs.FS, options GrepOptions) ([][]string, error) {
	var wg sync.WaitGroup
	var outputChans []chan string
	var results [][]string

	fs.WalkDir(fSys, options.path, func(path string, d fs.DirEntry, err error) error {
		outputChan := make(chan string)
		outputChans = append(outputChans, outputChan)

		wg.Add(1)
		go func(outputChan chan string) {
			defer wg.Done()
			defer close(outputChan)

			if err != nil {
				outputChan <- err.Error()
				return
			}

			if d.IsDir() {
				return
			}

			options := GrepOptions{path: path, stdin: nil, keyword: options.keyword, ignoreCase: options.ignoreCase, linesBeforeMatch: options.linesBeforeMatch}
			result, errGrep := Grep(fSys, options)
			if errGrep != nil {
				outputChan <- errGrep.Error()
				return
			}

			for _, r := range result {
				outputChan <- fmt.Sprintf("%s:%s", path, r)
			}
			
		}(outputChan)

		return nil
	})

	// A receive operation on a closed channel can always proceed immediately, 
	// yielding the element type's zero value after any previously sent values have been received.
	for _, outputChan := range outputChans {
		var result []string
		for resultStr := range outputChan {
			result = append(result, resultStr)
		}
		
		if len(result) > 0 {
			results = append(results, result)
		}
	}
	return results, nil
}