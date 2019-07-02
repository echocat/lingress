package rules

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"reflect"
	"testing"
)

func Test_ParsePath_works_as_expected(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect("/").To(beParsablePath(false))
	g.Expect("/a").To(beParsablePath(false, "a"))
	g.Expect("/a/").To(beParsablePath(false, "a"))
	g.Expect("/a/b").To(beParsablePath(false, "a", "b"))
	g.Expect("/a/b/").To(beParsablePath(false, "a", "b"))
	g.Expect("/a/b/c").To(beParsablePath(false, "a", "b", "c"))
	g.Expect("/a/b/c/").To(beParsablePath(false, "a", "b", "c"))

	g.Expect("").To(beParsablePath(true))
	g.Expect("//").To(beParsablePath(true, ""))
}

func Test_ParsePath_fails_as_expected(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect("").To(beNotParsablePath(false, ErrIllegalPath))
	g.Expect("//").To(beNotParsablePath(false, ErrIllegalPath))
	g.Expect("a").To(beNotParsablePath(false, ErrIllegalPath))
	g.Expect("a/b").To(beNotParsablePath(false, ErrIllegalPath))
	g.Expect("a/b/c").To(beNotParsablePath(false, ErrIllegalPath))
	g.Expect("a/").To(beNotParsablePath(false, ErrIllegalPath))
	g.Expect("a/b/").To(beNotParsablePath(false, ErrIllegalPath))
	g.Expect("a/b/c/").To(beNotParsablePath(false, ErrIllegalPath))

	g.Expect("a").To(beNotParsablePath(true, ErrIllegalPath))
	g.Expect("a/b").To(beNotParsablePath(true, ErrIllegalPath))
	g.Expect("a/b/c").To(beNotParsablePath(true, ErrIllegalPath))
}

func beParsablePath(faultTolerant bool, expected ...string) *shouldParsePathMatcher {
	if expected == nil {
		expected = []string{}
	}
	return &shouldParsePathMatcher{
		faultTolerant: faultTolerant,
		expected:      expected,
	}
}

func beNotParsablePath(faultTolerant bool, expectedError interface{}) *shouldFailParsePathMatcher {
	return &shouldFailParsePathMatcher{
		faultTolerant: faultTolerant,
		expectedError: expectedError,
	}
}

type shouldParsePathMatcher struct {
	faultTolerant bool
	expected      []string
	lastResult    []string
}

func (instance *shouldParsePathMatcher) Match(actual interface{}) (success bool, err error) {
	path, err := ParsePath(fmt.Sprint(actual), instance.faultTolerant)
	instance.lastResult = path
	if err != nil {
		return false, fmt.Errorf("'%v' should be parseable, but it's not; Got: %v", actual, err)
	}
	return reflect.DeepEqual(path, instance.expected), nil
}

func (instance *shouldParsePathMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%s\nBut got\n%s", format.Message(actual, "ParsePath should be equal with with", instance.expected), format.Object(instance.lastResult, 1))
}

func (instance *shouldParsePathMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%s\nBut got\n%s", format.Message(actual, "ParsePath should be not equal with with", instance.expected), format.Object(instance.lastResult, 1))
}

type shouldFailParsePathMatcher struct {
	faultTolerant bool
	expectedError interface{}
	lastError     error
}

func (instance *shouldFailParsePathMatcher) Match(actual interface{}) (success bool, err error) {
	_, err = ParsePath(fmt.Sprint(actual), instance.faultTolerant)
	if err != nil {
		if ee, ok := instance.expectedError.(string); ok {
			return err.Error() == ee, nil
		} else {
			return err == instance.expectedError, nil
		}
	}
	return false, nil
}

func (instance *shouldFailParsePathMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%s\nBut got\n%s", format.Message(actual, "ParsePath should fail with", instance.expectedError), format.Object(instance.lastError, 1))
}

func (instance *shouldFailParsePathMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%s\nBut got\n%s", format.Message(actual, "ParsePath should not fail with", instance.expectedError), format.Object(instance.lastError, 1))
}
