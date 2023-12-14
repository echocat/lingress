package tree

import "strings"

type testValue string

func (t testValue) Clone() testValue {
	return testValue(strings.Clone(string(t)))
}
