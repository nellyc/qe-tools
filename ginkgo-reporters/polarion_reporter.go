/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package ginkgo_reporters

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

var Polarion = PolarionReporter{}

func init() {
	flag.BoolVar(&Polarion.Run, "polarion-execution", false, "Run Polarion reporter")
	flag.StringVar(&Polarion.projectId, "polarion-project-id", "", "Set Polarion project ID")
	flag.StringVar(&Polarion.filename, "polarion-report-file", "polarion_results.xml", "Set Polarion report file path")
	flag.StringVar(&Polarion.plannedIn, "polarion-custom-plannedin", "", "Set Polarion planned-in ID")
    flag.StringVar(&Polarion.tier, "test-tier", "", "Set test tier number")
}

type PolarionTestSuite struct {
	XMLName      xml.Name           `xml:"testsuite"`
	Properties   PolarionProperties `xml:"properties"`
	TestCases    []PolarionTestCase `xml:"testcase"`
}

type PolarionTestCase struct {
	Name         String               `xml:"name,attr"`
}

type PolarionProperties struct {
	Property []PolarionProperty `xml:"property"`
}

type PolarionProperty struct {
	Name   string             `xml:"name,attr"`
	Value  string             `xml:"value,attr"`
}

type PolarionReporter struct {
	suite         PolarionTestSuite
	Run           bool
	filename      string
    testSuiteName string
}

func (reporter *PolarionReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {

	reporter.suite = PolarionTestSuite{
	    Properties: []PolarionProperties{},
		TestCases: []PolarionTestCase{},
	}

	property_map := make(map[string]string)
    property_map["polarion-project-id"] = reporter.projectId
    property_map["polarion-testcase-lookup-method"] = "name"
    property_map["polarion-custom-plannedin"] = "reporter.plannedIn"
    property_map["polarion-testrun-id"] = reporter.plannedIn+"_"+reporter.tier
    property_map["polarion-custom-isautomated"] = "True"

	properties := PolarionProperties{}
	for key, value := range property_map {
	    properties.Property = append(properties.Property, PolarionProperty{
            Name:   key,
            Value:  value,
	    })
	}

	reporter.suite.Properties = properties
	reporter.testSuiteName = summary.SuiteDescription
}

func (reporter *PolarionReporter) SpecWillRun(specSummary *types.SpecSummary) {
}

func (reporter *PolarionReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *PolarionReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *PolarionReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	testName := fmt.Sprintf(
		"%s: %s",
		specSummary.ComponentTexts[1],
		strings.Join(specSummary.ComponentTexts[2:], " "),
	)
	testCase := PolarionTestCase{
		Name:   testName,
	}

	reporter.suite.TestCases = append(reporter.suite.TestCases, testCase)
}

func (reporter *PolarionReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	if reporter.projectId == "" {
		fmt.Println("Can not create Polarion report without project ID")
		return
	}
	if reporter.plannedIn == "" {
        fmt.Println("Can not create Polarion report without planned-in ID")
        return
    }

	file, err := os.Create(reporter.filename)
	if err != nil {
		fmt.Printf("Failed to create Polarion report file: %s\n\t%s", reporter.filename, err.Error())
		return
	}
	defer file.Close()
	file.WriteString(xml.Header)
	encoder := xml.NewEncoder(file)
	encoder.Indent("  ", "    ")
	err = encoder.Encode(reporter.suite)
	if err != nil {
		fmt.Printf("Failed to generate Polarion report\n\t%s", err.Error())
	}
}
