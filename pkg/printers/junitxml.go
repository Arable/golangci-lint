package printers

import (
	"context"
	"encoding/xml"
	"io"
	"strings"

	"github.com/golangci/golangci-lint/pkg/result"
)

type testSuitesXML struct {
	XMLName    xml.Name `xml:"testsuites"`
	TestSuites []testSuiteXML
}

type testSuiteXML struct {
	XMLName   xml.Name      `xml:"testsuite"`
	Suite     string        `xml:"name,attr"`
	Tests     int           `xml:"tests,attr"`
	Errors    int           `xml:"errors,attr"`
	Failures  int           `xml:"failures,attr"`
	TestCases []testCaseXML `xml:"testcase"`
}

type testCaseXML struct {
	Name      string     `xml:"name,attr"`
	ClassName string     `xml:"classname,attr"`
	Failure   failureXML `xml:"failure"`
}

type failureXML struct {
	Message string `xml:"message,attr"`
	Content string `xml:",cdata"`
}

type JunitXML struct {
	w io.Writer
}

func NewJunitXML(w io.Writer) *JunitXML {
	return &JunitXML{w: w}
}

func (p JunitXML) Print(ctx context.Context, issues []result.Issue) error {
	suites := make(map[string]testSuiteXML) // use a map to group by file

	for ind := range issues {
		i := &issues[ind]
		suiteName := i.FilePath()
		testSuite := suites[suiteName]
		testSuite.Suite = i.FilePath()
		testSuite.Tests++
		testSuite.Failures++

		tc := testCaseXML{
			Name:      i.FromLinter,
			ClassName: i.Pos.String(),
			Failure: failureXML{
				Message: i.Text,
				Content: strings.Join(i.SourceLines, "\n"),
			},
		}

		testSuite.TestCases = append(testSuite.TestCases, tc)
		suites[suiteName] = testSuite
	}

	var res testSuitesXML
	for _, val := range suites {
		res.TestSuites = append(res.TestSuites, val)
	}

	enc := xml.NewEncoder(p.w)
	enc.Indent("", "  ")
	if err := enc.Encode(res); err != nil {
		return err
	}
	return nil
}
