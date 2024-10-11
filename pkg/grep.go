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

const MAX_OPEN_FILE_DESCRIPTORS = 1024

var (
	ErrIsDirectory = errors.New("is a directory")
)

type GrepOption struct {
	OrigPath         string
	Path             string
	Stdin            io.Reader
	Keyword          string
	IgnoreCase       bool
	LinesBeforeMatch int
	LinesAfterMatch  int
	SearchDir        bool
	LineCount        bool
}

type GrepResult struct {
	Path         string
	MatchedLines []string
	LineCount    int
	Error        error
}

func GrepR(fSys fs.FS, parentOption GrepOption) []GrepResult {
	var openFileLimit int = MAX_OPEN_FILE_DESCRIPTORS
	cond := sync.NewCond(&sync.Mutex{})

	var wg sync.WaitGroup
	var outputChans []chan GrepResult

	// walks over files in the directory
	fs.WalkDir(fSys, parentOption.Path, func(path string, d fs.DirEntry, err error) error {
		outputChan := make(chan GrepResult)
		outputChans = append(outputChans, outputChan)

		wg.Add(1)
		go func(outputChan chan GrepResult) {
			defer wg.Done()
			defer close(outputChan)

			if err != nil {
				outputChan <- GrepResult{Error: err}
				return
			}

			if d.IsDir() {
				return
			}

			// prepares the options for grep
			grepOption := GrepOption{
				Path:             path,
				OrigPath:         parentOption.Path,
				Keyword:          parentOption.Keyword,
				IgnoreCase:       parentOption.IgnoreCase,
				LinesBeforeMatch: parentOption.LinesBeforeMatch,
				LinesAfterMatch:  parentOption.LinesAfterMatch,
				LineCount:        parentOption.LineCount,
			}

			// goroutine will occupy the limit here if its available
			// otherwise it will wait
			cond.L.Lock()
			for openFileLimit <= 0 {
				cond.Wait()
			}
			openFileLimit--
			cond.L.Unlock()

			// grep operation here
			result := Grep(fSys, grepOption)
			if result.Error != nil {
				outputChan <- result
				return
			}

			// goroutine will free the limit here and signal other waiting goroutine to resume
			cond.L.Lock()
			openFileLimit++
			cond.Signal()
			cond.L.Unlock()

			// if no match found, then return
			if len(result.MatchedLines) == 0 && result.LineCount == 0 {
				return
			}

			// setting the path of file (from the user provided path)
			result.Path = normalisePathFromRoot(path, parentOption.OrigPath)
			outputChan <- result
		}(outputChan)

		return nil
	})

	var results []GrepResult // to save the final output
	// collates the results from all the output channels
	for _, outputChan := range outputChans {
		result := <-outputChan
		if len(result.MatchedLines) == 0 && result.LineCount == 0 {
			continue
		}
		results = append(results, result)
	}
	return results
}

func Grep(fSys fs.FS, option GrepOption) GrepResult {
	// gets the reader for file after validity checks
	r, cleanup, err := getReader(fSys, option)
	if err != nil {
		return GrepResult{Error: err}
	}
	defer cleanup()

	// searches for string
	result, err := searchString(r, option)
	if err != nil {
		return GrepResult{Error: err}
	}

	// prepares the result of string search
	res := GrepResult{
		Path: option.Path,
	}
	if option.LineCount {
		res.LineCount = len(result)
	} else {
		res.MatchedLines = result
	}

	return res
}

// gets reader for the file
func getReader(fSys fs.FS, option GrepOption) (io.Reader, func(), error) {
	if option.Path != "" {
		err := isValid(fSys, option)
		if err != nil {
			return nil, func() {}, err
		}

		file, err := fSys.Open(option.Path)
		if err != nil {
			return nil, nil, err
		}
		return file, func() { file.Close() }, nil
	}
	return option.Stdin, func() {}, nil
}

// main logic of string search
func searchString(r io.Reader, option GrepOption) ([]string, error) {
	// init buffer
	grepBuffer := NewGrepBuffer(option.LinesBeforeMatch)

	// counter for lines to save after match
	afterMatchCount := 0

	keyword := option.Keyword
	if option.IgnoreCase { // normalising keyword if ignoreCase was passed
		keyword = strings.ToLower(option.Keyword)
	}

	var result []string // to save final output
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		// normalising line if ignoreCase
		if option.IgnoreCase {
			line = strings.ToLower(scanner.Text())
		}

		// saves line in output
		// if match was found in prev iteration and user wants lines after match
		if afterMatchCount > 0 {
			result = append(result, scanner.Text())
			afterMatchCount--
		}

		// comparison and saving lines if matched
		if strings.Contains(line, keyword) {
			// saving lines if before match was passed
			if option.LinesBeforeMatch > 0 {
				result = append(result, grepBuffer.Dump()...)
			}

			// saving the matched line
			result = append(result, scanner.Text())

			// setting the counter for afterMatchCount if after match flag was passed
			if option.LinesAfterMatch > 0 {
				afterMatchCount += option.LinesAfterMatch
			}
		}

		// save line to buffer in advance
		// if match is found in future iteration and user wants lines before match
		if option.LinesBeforeMatch > 0 {
			grepBuffer.Push(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// checks if file is valid for reading
func isValid(fSys fs.FS, option GrepOption) error {
	// gets the file details
	fileInfo, err := fs.Stat(fSys, option.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", option.OrigPath, fs.ErrNotExist)
		}
		return fmt.Errorf("%s: %w", option.Path, err)
	}

	// checks for directory
	if fileInfo.IsDir() {
		return fmt.Errorf("%s: %w", option.OrigPath, ErrIsDirectory)
	}

	// checks for permissions
	// looks hacky, might have to change later
	if fileInfo.Mode().Perm()&400 == 0 {
		return fmt.Errorf("%s: %w", option.Path, fs.ErrPermission)
	}

	return nil
}

// returns the file path from user provided path
func normalisePathFromRoot(rootPath, dirPath string) string {
	dirPathClean := strings.TrimPrefix(dirPath, "../")
	idx := strings.Index(rootPath, dirPathClean)

	return dirPath + rootPath[idx+len(dirPathClean):]
}
