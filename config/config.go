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
	// DeprecatedTestCaseIDPrefix refers to the deprecated prefix
	// associated with test case ids.
	DeprecatedTestCaseIDPrefix string = "tcid-"

	// ActualTestCaseNamePrefix refers to the prefix associated with
	// every test case name
	//
	// NOTE:
	//	Actual test cases are ones whose testcases are found at
	// .gitlab-ci.yml
	ActualTestCaseNamePrefix string = "TCID-"

	// ActualTestCaseNameDelimiter is the delimiter used to separate
	// the testcase name from its prefix & test case id value & actual
	// testcase name
	//
	// NOTE:
	//	tcid-miot1x-maya-io-server-check: is the yaml line where:
	//	- tcid is the prefix
	//	- miot1x is the tcid value
	//	- maya-io-server-check is the testcase name
	ActualTestCaseNameDelimiter string = "-"

	// DesiredTestCaseNamePrefix refers to the prefix associated with
	// every test case name
	//
	// NOTE:
	//	Desired test cases are ones whose testcases are found at
	// .master-plan.yml
	DesiredTestCaseNamePrefix string = "- tcid:"

	// DesiredTestCaseNameDelimiter is the delimiter used to separate
	// the testcase name from its prefix & test case id value
	//
	// NOTE:
	//	tcid: miot1x is the yaml line where:
	//	- tcid is the prefix
	//	- miot1x is the tcid value
	DesiredTestCaseNameDelimiter string = ": "
)

// MetricsConfig has required details on actual vs. desired
// e2e test cases
type MetricsConfig struct {
	DesiredTestCases    map[string]bool
	ActualTestCases     map[string]bool
	DeprecatedTestCases []string
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
		DesiredTestCasesFileName: ".master-plan.yml",
		ActualTestCasesFileName:  ".gitlab-ci.yml",
	}
}

// LoadOrEmpty loads all config files if available or
// returns empty config along with load error
func (c *Config) LoadOrEmpty() (*MetricsConfig, error) {
	mc, err := c.Load()
	if err != nil {
		mc = &MetricsConfig{
			DesiredTestCases: map[string]bool{},
			ActualTestCases:  map[string]bool{},
		}
	}
	return mc, err
}

// Load loads all config files or load error
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

	registerActualTestCaseNamesFn := func(lineContent string) {
		glog.V(6).Infof("Actual: Received %q", lineContent)
		lineContent = strings.TrimSpace(lineContent)
		if strings.HasPrefix(lineContent, ActualTestCaseNamePrefix) {
			tcid := strings.TrimSuffix(lineContent, ":")
			glog.V(3).Infof("Registering actual tcid %q", tcid)
			out.ActualTestCases[tcid] = true
		} else if strings.HasPrefix(lineContent, DeprecatedTestCaseIDPrefix) {
			tcid := strings.TrimSuffix(lineContent, ":")
			glog.V(3).Infof("Registering deprecated tcid %q", tcid)
			out.DeprecatedTestCases = append(
				out.DeprecatedTestCases,
				tcid,
			)
		}
	}
	registerDesiredTestCaseNamesFn := func(lineContent string) {
		glog.V(6).Infof("Desired: Received %q", lineContent)
		lineContent = strings.TrimSpace(lineContent)
		if strings.HasPrefix(lineContent, DesiredTestCaseNamePrefix) {
			words := strings.Split(lineContent, DesiredTestCaseNameDelimiter)
			if len(words) == 2 {
				tcid := strings.TrimSpace(words[1])
				glog.V(3).Infof("Registering desired tcid %q", tcid)
				out.DesiredTestCases[tcid] = true
			}
		}
	}
	// we support e2e metrics yaml files only
	getRegisterTestCaseNamesFuncForFileName :=
		func(filename string) func(string) {
			if filename == c.ActualTestCasesFileName {
				return registerActualTestCaseNamesFn
			} else if filename == c.DesiredTestCasesFileName {
				return registerDesiredTestCaseNamesFn
			}
			return nil
		}
	// there will be multiple config files i.e. desired & actual test
	// case files in one location
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() || file.Mode().IsDir() {
			glog.V(4).Infof(
				"Will skip config %q at path %q: Not a file",
				fileName,
				c.Path,
			)
			// we don't load folder(s)
			continue
		}
		if !strings.HasSuffix(fileName, ".yaml") &&
			!strings.HasSuffix(fileName, ".yml") {
			glog.V(4).Infof(
				"Will skip config %q at path %q: Not a yaml file",
				fileName,
				c.Path,
			)
			// we support only yaml files
			continue
		}
		// get registry logic that registers all the test cases
		// found in this file based on the name of this file
		registerTestCaseNames := getRegisterTestCaseNamesFuncForFileName(fileName)
		if registerTestCaseNames == nil {
			glog.V(4).Infof(
				"Will skip config %q at path %q: Want %q or %q",
				fileName,
				c.Path,
				c.DesiredTestCasesFileName,
				c.ActualTestCasesFileName,
			)
			continue
		}
		// build full file path
		fileNameWithPath := c.Path + fileName
		glog.V(2).Infof("Will load config %q", fileNameWithPath)

		// logic that parses the file & registers the test case names
		// found in this file
		err := parseFileByLine(fileNameWithPath, registerTestCaseNames)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Failed to parse %q",
				fileNameWithPath,
			)
		}
	}
	glog.V(4).Infof("Config(s) loaded successfully from path %q", c.Path)
	return out, nil
}

// processFileByLine parses the given file using
// the provided parse logic
func parseFileByLine(filename string, process func(string)) (err error) {
	glog.V(3).Infof("Will parse file %q", filename)
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
	glog.V(4).Infof("Parsed file %q successfully", filename)
	return nil
}
