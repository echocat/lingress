package tree

import (
	. "github.com/onsi/gomega"
	"testing"
)

func Test_New_works_as_expected(t *testing.T) {
	g := NewGomegaWithT(t)

	actual := New()

	g.Expect(actual).ToNot(BeNil())
	g.Expect(actual.root).NotTo(BeNil())
	g.Expect(actual.root.elements).To(BeNil())
	g.Expect(actual.root.children).To(BeNil())
}
