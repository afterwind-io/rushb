package rushb

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
)

// Suite the test suite object
type Suite struct {
	T      *testing.T
	indent int
	passed int
	failed int
	skiped int
}

type suiteError struct {
	critical bool
	msg      string
}

// SuiteTestHandler defines a test wrapper function.
// If the return value is an error, the correspoding test
// will be marked failed.
type SuiteTestHandler = func(s *Suite) error

// NewSuite returns a reference to a fresh test suite
func NewSuite(t *testing.T) *Suite {
	// HACK: force color prompt, handy for some special use
	// cases(CI console, etc.)
	color.NoColor = false

	return &Suite{T: t}
}

// Start literally starts the test. Similar to `Title`,
// `Start` gives a heading, and outputs some
// basic statistics at the end of the test.
func (s *Suite) Start(name string, test func()) {
	defer func() {
		s.println("\n=== FINISHED\n")
		s.println(fmt.Sprintf("%v: %v", color.HiGreenString("Passed"), +s.passed))
		s.println(fmt.Sprintf("%v: %v", color.HiRedString("Failed"), +s.failed))
		s.println(fmt.Sprintf("%v: %v", color.HiBlueString("Skiped"), +s.skiped))
		s.println(fmt.Sprintf("\nTotal: %v\n", s.passed+s.failed+s.skiped))
	}()

	s.Title(name, test)
}

// Title gives a summary to a group of tests.
// It is mainly for organizing purpose, to provide
// the results a better structure.
func (s *Suite) Title(name string, test func()) {
	defer func() {
		s.indent = s.indent - 2
	}()

	s.println("")
	s.title(name)
	s.println("")

	s.indent = s.indent + 2

	test()
}

// Check runs a single test. If the test call returns an error,
// this single test is marked failed and remaining tests continue.
// If the test call panics other than calling `Fatal`,
// the whole tests stop and exit.
func (s *Suite) Check(name string, test SuiteTestHandler) {
	defer func() {
		if err := recover(); err != nil {
			s.failed++
			s.fail(name)

			switch e := err.(type) {
			case suiteError:
				if e.critical {
					s.T.Fatal(e.msg)
				} else {
					s.T.Error(e.msg)
				}

			default:
				s.T.Error(err)
			}
		}
	}()

	if err := test(s); err != nil {
		s.failed++
		s.fail(name)
		s.T.Error(err)
	}

	s.passed++
	s.ok(name)
}

// Critical runs a single test. Unlike method "Check",
// if the test call returns an error, or panics,
// the whole tests fail and exit anyway.
func (s *Suite) Critical(name string, test SuiteTestHandler) {
	fail := func(err interface{}) {
		s.failed++
		s.fail(name)
		s.T.Fatal(err)
	}

	defer func() {
		if err := recover(); err != nil {
			fail(err)
		}
	}()

	if err := test(s); err != nil {
		fail(err)
	}

	s.passed++
	s.ok(name)
}

// Skip marks the test should be done later
func (s *Suite) Skip(name string, test SuiteTestHandler) {
	s.skiped++
	s.skip(name)
}

// Try is a handy function to test whether the specific
// function will panic or not. A `catch` callback can be provided
// to deal with the error.
func (s *Suite) Try(try func(), catch func(err interface{})) (result bool) {
	defer func() {
		if err := recover(); err != nil {
			if catch != nil {
				catch(err)
			}
			result = false
		} else {
			result = true
		}
	}()

	try()
	return result
}

// Assert defines a simple equality test, "==" inside.
func (s *Suite) Assert(actual interface{}, expect interface{}) {
	if actual != expect {
		panic(fmt.Sprintf(`Expect "%v", got "%v"`, expect, actual))
	}
}

// Info outputs some texts in grey color, similar to `Println`
func (s *Suite) Info(content string) {
	s.info(content)
}

// Fail marks current test failed, then continues execution
func (s *Suite) Fail(content string) {
	panic(suiteError{critical: false, msg: content})
}

// Fatal marks current test failed, then skips the remaining
// tests and exits.
func (s *Suite) Fatal(content string) {
	panic(suiteError{critical: true, msg: content})
}

func (s *Suite) title(content string) {
	c := color.New(color.Bold)
	s.println(c.Sprintf(content))
}

func (s *Suite) ok(content string) {
	s.println("[" + color.HiGreenString("Done") + "] " + content)
}

func (s *Suite) fail(content string) {
	s.println("[" + color.HiRedString("Fail") + "] " + content)
}

func (s *Suite) skip(content string) {
	s.println("[" + color.HiBlueString("Skip") + "] " + color.HiBlackString(content))
}

func (s *Suite) info(content string) {
	s.println(color.HiBlackString(content))
}

func (s *Suite) println(content string) {
	prefix := strings.Repeat(" ", s.indent)

	fmt.Println(prefix + content)
}
