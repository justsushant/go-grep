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
	OrigPath string
	Path string
	Stdin io.Reader
	Keyword string
	FileWName string
	IgnoreCase bool
	LinesBeforeMatch int
	LinesAfterMatch int
	SearchDir bool
	LineCount bool
}

type GrepResult struct {
	Path string
	MatchedLines []string
	LineCount int
	Error error
}

func GrepR(fSys fs.FS, parentOption GrepOptions) []GrepResult {
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
			grepOption := GrepOptions{
				Path: path, 
				OrigPath: parentOption.Path, 
				Keyword: parentOption.Keyword, 
				IgnoreCase: parentOption.IgnoreCase, 
				LinesBeforeMatch: parentOption.LinesBeforeMatch, 
				LinesAfterMatch: parentOption.LinesAfterMatch, 
				LineCount: parentOption.LineCount,
			}
			result := Grep(fSys, grepOption)
			if result.Error != nil {
				outputChan <- result
				return
			}
			
			// if no match found, then return
			if len(result.MatchedLines) == 0 && result.LineCount == 0 {
				return
			}

			// setting the path of file (from the user provided path)
			result.Path = normalisePathFromRoot(path, parentOption.OrigPath)
			outputChan <- result
		} (outputChan)

		return nil
	})

	var results []GrepResult	// to save the final output
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

func Grep(fSys fs.FS, option GrepOptions) GrepResult {
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
func getReader(fSys fs.FS, option GrepOptions) (io.Reader, func(), error) {
	if option.Path != "" {
		err := isValid(fSys, option.Path, option.OrigPath)
		if err != nil {
			return nil, nil, err
		}

		file, err := fSys.Open(option.Path)
		if err != nil {
			return nil, nil, err
		}
		return file, func() {file.Close()}, nil
	}
	return option.Stdin, func() {}, nil
}

// main logic of string search
func searchString(r io.Reader, options GrepOptions) ([]string, error) {
	// init buffer
	grepBuffer := NewGrepBuffer(options.LinesBeforeMatch)	
	// counter for lines to save after match
	afterMatchCount := 0
	
	keyword := options.Keyword
	if options.IgnoreCase {		// normalising keyword if ignoreCase was passed
		keyword = strings.ToLower(options.Keyword)
	}

	var result []string		// to save final output
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		
		// saves lines after match in output
		if afterMatchCount > 0 {
			result = append(result, scanner.Text())
			afterMatchCount--
		}

		// normalising line if ignoreCase
		if options.IgnoreCase {
			line = strings.ToLower(scanner.Text())
		}

		// comparison and saving lines if matched
		if strings.Contains(line, keyword) {
			// saving lines if before match was passed
			if options.LinesBeforeMatch > 0 {
				result = append(result, grepBuffer.Dump()...)
			}

			// saving the matched line
			result = append(result, scanner.Text())
			
			// saving lines if after match was passed
			if options.LinesAfterMatch > 0 {
				afterMatchCount = options.LinesAfterMatch
			}
		}
		
		// save lines to buffer
		if options.LinesBeforeMatch > 0 {
			grepBuffer.Push(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// checks if file is valid for reading
func isValid(fSys fs.FS, path, origPath string) error {
	// gets the file details
	fileInfo, err := fs.Stat(fSys, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", origPath, fs.ErrNotExist)
		}
		return fmt.Errorf("%s: %w", path, err)
	}

	// checks for directory
	if fileInfo.IsDir() {
		return fmt.Errorf("%s: %w", origPath, ErrIsDirectory)
	}

	// checks for permissions
	// looks hacky, might have to change later
	if fileInfo.Mode().Perm()&400 == 0 {
		return fmt.Errorf("%s: %w", path, fs.ErrPermission)
	}

	return nil
}

// returns the file path from user provided path
func normalisePathFromRoot(rootPath, userPath string) string {
	userPathClean := strings.TrimPrefix(userPath, "../")
    idx := strings.Index(rootPath, userPathClean)

	return userPath + rootPath[idx+len(userPathClean):]
}