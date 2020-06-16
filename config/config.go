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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	prom "mayadata.io/e2e-metrics/metrics"
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

// TestCasesMetrics has required details on actual vs. desired
// e2e test cases
type TestCasesMetrics struct {
	DesiredTestCases    map[string]bool
	ActualTestCases     map[string]bool
	DeprecatedTestCases []string
}

// Loadable helps loading the testcase config files
type Loadable struct {
	log  logr.Logger
	prom *prom.Metrics
	Path string

	// File that has all the desired test cases
	DesiredTestCasesFileName string

	// File that has the implemented test cases
	ActualTestCasesFileName string
}

type LoadableConfig struct {
	Log  logr.Logger
	Prom *prom.Metrics
	Path string
}

// New returns a new instance of config
func New(conf LoadableConfig) *Loadable {
	return &Loadable{
		Path:                     conf.Path,
		log:                      conf.Log,
		prom:                     conf.Prom,
		DesiredTestCasesFileName: ".master-plan.yml",
		ActualTestCasesFileName:  ".gitlab-ci.yml",
	}
}

// LoadOrEmpty loads all config files if available or
// returns empty config along with load error
func (c *Loadable) LoadOrEmpty() (*TestCasesMetrics, error) {
	mc, err := c.Load()
	if err != nil {
		// set an empty metrics if error
		mc = &TestCasesMetrics{
			DesiredTestCases: map[string]bool{},
			ActualTestCases:  map[string]bool{},
		}
	}
	return mc, err
}

// Load loads all config files or load error
func (c *Loadable) Load() (*TestCasesMetrics, error) {
	log := c.log
	log.V(3).Info("Will load test case config(s)", "path", c.Path)

	files, readDirErr := ioutil.ReadDir(c.Path)
	if readDirErr != nil {
		return nil, readDirErr
	}
	if len(files) == 0 {
		return nil,
			errors.Errorf("No config(s) found at %q", c.Path)
	}

	var out = &TestCasesMetrics{
		DesiredTestCases: map[string]bool{},
		ActualTestCases:  map[string]bool{},
	}

	registerActualTestCaseNamesFn := func(lineContent string) {
		log.V(6).Info("Actual testcase", "Received", lineContent)
		lineContent = strings.TrimSpace(lineContent)
		if strings.HasPrefix(lineContent, ActualTestCaseNamePrefix) {
			tcid := strings.TrimSuffix(lineContent, ":")
			log.V(3).Info("Registering actual tcid", "name", tcid)
			out.ActualTestCases[tcid] = true
		} else if strings.HasPrefix(lineContent, DeprecatedTestCaseIDPrefix) {
			tcid := strings.TrimSuffix(lineContent, ":")
			log.V(3).Info("Registering deprecated tcid", "name", tcid)
			out.DeprecatedTestCases = append(
				out.DeprecatedTestCases,
				tcid,
			)
		}
	}
	registerDesiredTestCaseNamesFn := func(lineContent string) {
		log.V(6).Info("Desired testcase", "Received", lineContent)
		lineContent = strings.TrimSpace(lineContent)
		if strings.HasPrefix(lineContent, DesiredTestCaseNamePrefix) {
			words := strings.Split(lineContent, DesiredTestCaseNameDelimiter)
			if len(words) == 2 {
				tcid := strings.TrimSpace(words[1])
				log.V(3).Info("Registering desired tcid", "name", tcid)
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
			log.V(4).Info(
				"Will skip config: Not a file",
				"file", fileName,
				"path", c.Path,
			)
			// we don't load folder(s)
			continue
		}
		if !strings.HasSuffix(fileName, ".yaml") &&
			!strings.HasSuffix(fileName, ".yml") {
			log.V(4).Info(
				"Will skip config: Not a yaml file",
				"file", fileName,
				"path", c.Path,
			)
			// we support only yaml files
			continue
		}
		// get registry logic that registers all the test cases
		// found in this file based on the name of this file
		registerTestCaseNames := getRegisterTestCaseNamesFuncForFileName(fileName)
		if registerTestCaseNames == nil {
			log.V(4).Info(
				"Will skip config",
				"got-file", fileName,
				"path", c.Path,
				"want-file", c.DesiredTestCasesFileName, c.ActualTestCasesFileName,
			)
			continue
		}
		// build full file path
		fileNameWithPath := c.Path + fileName
		log.V(2).Info("Will load config", "file", fileNameWithPath)

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
	log.V(4).Info("Config(s) loaded successfully", "path", c.Path)

	actualTestCaseCount := len(out.ActualTestCases)
	desiredTestCaseCount := len(out.DesiredTestCases)
	c.prom.SetActualTestCount(&prom.ActualTestCount{
		BaseTestCount: prom.BaseTestCount{
			Value:                  float64(actualTestCaseCount),
			TestImplementationType: prom.TestImplementationTypeLitmus,
		},
	})
	c.prom.SetPlannedTestCount(&prom.PlannedTestCount{
		BaseTestCount: prom.BaseTestCount{
			Value:                  float64(desiredTestCaseCount),
			TestImplementationType: prom.TestImplementationTypeLitmus,
		},
	})
	log.V(4).Info(
		"Prometheus metrics were set",
		"actual-test-count", actualTestCaseCount,
		"desired-test-count", desiredTestCaseCount,
	)
	return out, nil
}

// processFileByLine parses the given file using
// the provided parse logic
func parseFileByLine(filename string, process func(string)) (err error) {
	klog.V(3).Infof("Will parse file %q", filename)
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
	klog.V(4).Infof("Parsed file %q successfully", filename)
	return nil
}
