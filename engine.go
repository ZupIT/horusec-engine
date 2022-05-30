// Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/ZupIT/horusec-engine/pool"
)

// AcceptAnyExtension can be passed as extensions argument in NewEngine to accept any extension
const AcceptAnyExtension string = "*"

// Finding represents a possible vulnerability found by the engine, it contains all information necessary to detect and
// correct the vulnerability
type Finding struct {
	ID             string
	Name           string
	Severity       string
	CodeSample     string
	Confidence     string
	Description    string
	SourceLocation Location
}

// Location represents the location of the vulnerability in a file
type Location struct {
	Filename string
	Line     int
	Column   int
}

// Engine contains all the engine necessary data
type Engine struct {
	poolSize   int
	extensions []string
}

// NewEngine creates a new engine instance with all necessary data.
// extensions argument represents which extension the engine should apply the rules
// poolSize represents the number of go routines to open (Default is 10)
func NewEngine(poolSize int, extensions ...string) *Engine {
	return &Engine{
		poolSize:   poolSize,
		extensions: extensions,
	}
}

// Run walks through projectPath and runs the method Rule.Run in a pool of goroutines
// if an error is found when executes Rule.Run method it cancels current running go routines and return
// valid findings and the error
// nolint:funlen,gocyclo // necessary complexity, breaking this function will lead to an even more complex code
func (e *Engine) Run(ctx context.Context, projectPath string, rules ...Rule) ([]Finding, error) {
	var findings []Finding

	paths, err := e.getValidFilePaths(projectPath)
	if err != nil {
		return nil, err
	}

	mutex := new(sync.Mutex)
	wg := sync.WaitGroup{}

	workerPool, err := pool.NewPool(e.poolSize)
	if err != nil {
		return nil, err
	}

	defer workerPool.Release()

	group, _ := errgroup.WithContext(ctx)

	wg.Add(len(paths))

	for _, path := range paths {
		pathCopy := path

		errSubmit := workerPool.Submit(func() {
			group.Go(func() error {
				defer wg.Done()

				newFindings, errRunRule := e.runRule(rules, pathCopy)
				if errRunRule != nil {
					return errRunRule
				}

				mutex.Lock()
				findings = append(findings, newFindings...)
				mutex.Unlock()

				return errRunRule
			})
		})
		if errSubmit != nil {
			return nil, errSubmit
		}
	}

	wg.Wait()
	err = group.Wait()

	return findings, err
}

func (e *Engine) runRule(rules []Rule, pathCopy string) ([]Finding, error) {
	var findings []Finding

	for _, rule := range rules {
		f, err := rule.Run(pathCopy)
		if err != nil {
			return nil, err
		}

		findings = append(findings, f...)
	}

	return findings, nil
}

// getValidFilePaths this function will walk the project directory and will look for files that match the extensions
// informed during the initialization of the engine and return a slice with it.
// Directories, sys links and files with extensions that are not in Engine.extensions struct wil be ignored
func (e *Engine) getValidFilePaths(projectPath string) ([]string, error) {
	var validPaths []string

	err := filepath.WalkDir(projectPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || e.isInvalidFilePath(path, entry) {
			return err
		}

		validPaths = append(validPaths, path)

		return nil
	})

	return validPaths, err
}

// isInvalidFilePath contains a list of validations to check if a path needs to be analyzed. It will ignore directories,
// sysLinks, extensions that don't match the necessary ones, and .git files
func (e *Engine) isInvalidFilePath(path string, entry fs.DirEntry) bool {
	return entry.IsDir() ||
		entry.Type() == fs.ModeSymlink ||
		e.isInvalidExtension(path) ||
		e.isFileFromGitFolder(path)
}

// isInvalidExtension verify if the filepath contains a valid file extension.
// The valid extensions are the ones that should be analyzed, and are passed during the initialization of the engine
func (e *Engine) isInvalidExtension(path string) bool {
	for _, ext := range e.extensions {
		if ext == filepath.Ext(path) || ext == AcceptAnyExtension {
			return false
		}
	}

	return true
}

// isFileFromGitFolder check if a file is in a .git folder
func (e *Engine) isFileFromGitFolder(path string) bool {
	return strings.Contains(path, ".git"+string(os.PathSeparator))
}
