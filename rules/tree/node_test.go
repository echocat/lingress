package tree

import (
	. "github.com/onsi/gomega"
	"testing"
)

func Test_newNode_works_as_expected(t *testing.T) {
	g := NewGomegaWithT(t)

	actual := newNode[testValue]()

	g.Expect(actual).ToNot(BeNil())
	g.Expect(actual.elements).To(BeNil())
	g.Expect(actual.children).To(BeNil())
}
