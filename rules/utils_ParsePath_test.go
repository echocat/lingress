package rules

import (
	"errors"
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

func beNotParsablePath(faultTolerant bool, expectedError any) *shouldFailParsePathMatcher {
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

func (this *shouldParsePathMatcher) Match(actual any) (success bool, err error) {
	path, err := ParsePath(fmt.Sprint(actual), this.faultTolerant)
	this.lastResult = path
	if err != nil {
		return false, fmt.Errorf("'%v' should be parseable, but it's not; Got: %v", actual, err)
	}
	return reflect.DeepEqual(path, this.expected), nil
}

func (this *shouldParsePathMatcher) FailureMessage(actual any) (message string) {
	return fmt.Sprintf("%s\nBut got\n%s", format.Message(actual, "ParsePath should be equal with with", this.expected), format.Object(this.lastResult, 1))
}

func (this *shouldParsePathMatcher) NegatedFailureMessage(actual any) (message string) {
	return fmt.Sprintf("%s\nBut got\n%s", format.Message(actual, "ParsePath should be not equal with with", this.expected), format.Object(this.lastResult, 1))
}

type shouldFailParsePathMatcher struct {
	faultTolerant bool
	expectedError any
	lastError     error
}

func (this *shouldFailParsePathMatcher) Match(actual any) (success bool, err error) {
	_, err = ParsePath(fmt.Sprint(actual), this.faultTolerant)
	if err != nil {
		if ee, ok := this.expectedError.(string); ok {
			return err.Error() == ee, nil
		} else {
			return errors.Is(err, this.expectedError.(error)), nil
		}
	}
	return false, nil
}

func (this *shouldFailParsePathMatcher) FailureMessage(actual any) (message string) {
	return fmt.Sprintf("%s\nBut got\n%s", format.Message(actual, "ParsePath should fail with", this.expectedError), format.Object(this.lastError, 1))
}

func (this *shouldFailParsePathMatcher) NegatedFailureMessage(actual any) (message string) {
	return fmt.Sprintf("%s\nBut got\n%s", format.Message(actual, "ParsePath should not fail with", this.expectedError), format.Object(this.lastError, 1))
}
