/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

const (
	// ActualTestCaseNamePrefix refers to the prefix associated with
	// every test case name
	ActualTestCaseNamePrefix string = "tcid-"

	// ActualTestCaseNameDelimiter is the delimiter used to separate
	// the testcase name from its prefix & test case id value & actual
	// testcase name
	//
	// NOTE:
	//	tcid-miot1x-maya-io-server-check: is the yaml line where:
	//	- tcid is a prefix
	//	- miot1x is the tcid value
	//	- maya-io-server-check is the testcase name
	ActualTestCaseNameDelimiter string = "-"

	// DesiredTestCaseNamePrefix refers to the prefix associated with
	// every test case name
	DesiredTestCaseNamePrefix string = "tcid:"

	// DesiredTestCaseNameDelimiter is the delimiter used to separate
	// the testcase name from its prefix & test case id value
	//
	// NOTE:
	//	tcid: miot1x is the yaml line where:
	//	- tcid is the field
	//	- miot1x is the tcid value
	DesiredTestCaseNameDelimiter string = ": "
)

// MetricsConfig has required details on actual vs. desired
// e2e test cases
type MetricsConfig struct {
	DesiredTestCases map[string]bool
	ActualTestCases  map[string]bool
}

// Config is the path to the config files
type Config struct {
	Path string

	// File that has all the desired test cases
	DesiredTestCasesFileName string

	// File that has the implemented test cases
	ActualTestCasesFileName string
}

// New returns a new instance of config
func New(path string) *Config {
	return &Config{
		Path:                     path,
		DesiredTestCasesFileName: "master-ci.yml",
		ActualTestCasesFileName:  ".gitlab-ci.yml",
	}
}

// Load loads all metac config files & converts them
// to unstructured instances
func (c *Config) Load() (*MetricsConfig, error) {
	glog.V(3).Infof("Will load config(s) from path %q", c.Path)

	files, readDirErr := ioutil.ReadDir(c.Path)
	if readDirErr != nil {
		return nil, readDirErr
	}
	if len(files) == 0 {
		return nil,
			errors.Errorf("No config(s) found at %q", c.Path)
	}

	var out = &MetricsConfig{
		DesiredTestCases: map[string]bool{},
		ActualTestCases:  map[string]bool{},
	}

	setActualTestCaseName := func(lineContent string) {
		glog.V(4).Infof("Actual: Received %q", lineContent)
		lineContent = strings.TrimSpace(lineContent)
		if strings.HasPrefix(lineContent, ActualTestCaseNamePrefix) {
			words := strings.Split(lineContent, ActualTestCaseNameDelimiter)
			if len(words) > 2 {
				tcid := words[1]
				glog.V(2).Infof("Adding actual tcid %q", tcid)
				out.ActualTestCases[tcid] = true
			}
		}
	}
	setDesiredTestCaseName := func(lineContent string) {
		glog.V(4).Infof("Desired: Received %q", lineContent)
		lineContent = strings.TrimSpace(lineContent)
		if strings.HasPrefix(lineContent, DesiredTestCaseNamePrefix) {
			words := strings.Split(lineContent, DesiredTestCaseNameDelimiter)
			if len(words) == 2 {
				tcid := words[1]
				glog.V(2).Infof("Adding desired tcid %q", tcid)
				out.DesiredTestCases[tcid] = true
			}
		}
	}
	setTestCaseName := func(filename string) func(string) {
		if c.ActualTestCasesFileName == filename {
			return setActualTestCaseName
		} else if c.DesiredTestCasesFileName == filename {
			return setDesiredTestCaseName
		}
		return nil
	}
	// there will be multiple config files i.e. desired & actual
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() || file.Mode().IsDir() {
			glog.V(3).Infof(
				"Will skip config %q at path %q: Not a file", fileName, c.Path,
			)
			// we don't load folder(s)
			continue
		}
		if !strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, ".yml") {
			glog.V(3).Infof(
				"Will skip config %q at path %q: Not a yaml file", fileName, c.Path,
			)
			// we support only yaml files
			continue
		}
		// get appropriate setter function
		setTestCaseNamesByLine := setTestCaseName(fileName)
		if setTestCaseNamesByLine == nil {
			glog.V(3).Infof(
				"Will skip config %q at path %q: Want %q or %q",
				fileName, c.Path, c.DesiredTestCasesFileName, c.ActualTestCasesFileName,
			)
			// we support only e2e metrics yaml files
			continue
		}
		// build full file path
		fileNameWithPath := c.Path + fileName
		glog.V(2).Infof("Will load config %q", fileNameWithPath)

		// real logic to load test case names
		err := processFileByLine(fileNameWithPath, setTestCaseNamesByLine)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to load %q", fileNameWithPath)
		}
	}
	glog.V(4).Infof("Config(s) loaded successfully from path %q", c.Path)
	return out, nil
}

func processFileByLine(filename string, process func(string)) (err error) {
	glog.V(3).Infof("Will process file %q", filename)
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)
	// loop over all lines in the file
	for {
		var linecontent bytes.Buffer
		var chunk []byte
		var isPrefix bool
		// loop over the buffer chunks inside one line
		for {
			chunk, isPrefix, err = reader.ReadLine()
			linecontent.Write(chunk)
			// If we've reached the end of the line
			if !isPrefix {
				// stop reading this line further
				break
			}
			// If we're just at the EOF, break
			if err != nil {
				break
			}
		}
		// if we have completed reading the entire file then stop
		if err == io.EOF {
			break
		}
		if err == nil {
			// at this point, the entire line's content will be available
			line := linecontent.String()
			// do not process a blank line
			if line != "" {
				// process the line in the provided callback fn
				process(line)
			}
		}
	}
	if err != io.EOF {
		return err
	}
	glog.V(4).Infof("Successfully processed file %q", filename)
	return nil
}
