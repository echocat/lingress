package tree

import (
	. "github.com/onsi/gomega"
	"strings"
	"testing"
)

func Test_Node_Put_works_with_element_in_root(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "a", element: "elementA"},
	}, node{
		elements: map[string][]interface{}{
			"a": {"elementA"},
		},
	}, nil)
}

func Test_Node_Put_works_with_element_in_subPath(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "a/b", element: "elementA"},
	}, node{
		children: map[string]*node{
			"a": {
				elements: map[string][]interface{}{
					"b": {"elementA"},
				},
			},
		},
	}, nil)
}

func Test_Node_Put_works_with_3_elements_in_root(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "", element: "elementRoot1"},
		{path: "", element: "elementRoot2"},
		{path: "", element: "elementRoot3"},
		{path: "a", element: "elementA"},
		{path: "b", element: "elementB"},
		{path: "c", element: "elementC"},
	}, node{
		elements: map[string][]interface{}{
			"a": {"elementA"},
			"b": {"elementB"},
			"c": {"elementC"},
		},
	}, &[]interface{}{"elementRoot1", "elementRoot2", "elementRoot3"})
}

func Test_Node_Put_works_with_3_elements_across_path(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "a", element: "elementA"},
		{path: "a/b", element: "elementAB"},
		{path: "a/b/c", element: "elementABC"},
	}, node{
		children: map[string]*node{
			"a": {
				children: map[string]*node{
					"b": {
						elements: map[string][]interface{}{
							"c": {"elementABC"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b": {"elementAB"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a": {"elementA"},
		},
	}, nil)
}

func Test_Node_Put_works_with_2_elements_across_path(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "a", element: "elementA"},
		{path: "a/b/c", element: "elementABC"},
	}, node{
		children: map[string]*node{
			"a": {
				children: map[string]*node{
					"b": {
						elements: map[string][]interface{}{
							"c": {"elementABC"},
						},
					},
				},
			},
		},
		elements: map[string][]interface{}{
			"a": {"elementA"},
		},
	}, nil)
}

func Test_Node_Put_works_with_same_elements_across_path(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "a", element: "elementA!"},
		{path: "a/b/c", element: "elementABC!"},
		{path: "a", element: "elementA"},
		{path: "a/b/c", element: "elementABC"},
	}, node{
		children: map[string]*node{
			"a": {
				children: map[string]*node{
					"b": {
						elements: map[string][]interface{}{
							"c": {"elementABC!", "elementABC"},
						},
					},
				},
			},
		},
		elements: map[string][]interface{}{
			"a": {"elementA!", "elementA"},
		},
	}, nil)
}

func executeTestPutRun(t *testing.T, given []putTestCaseGiven, expected node, expectedRootElements *[]interface{}) {
	g := NewGomegaWithT(t)

	actual := New()
	for _, gElement := range given {
		pathElements := strings.Split(gElement.path, "/")
		if gElement.path == "" {
			pathElements = []string{}
		}
		err := actual.Put(pathElements, gElement.element)
		g.Expect(err).To(BeNil())
	}

	g.Expect(actual.root).To(Equal(&expected))
	g.Expect(actual.rootElements).To(Equal(expectedRootElements))
}

type putTestCaseGiven struct {
	path    string
	element interface{}
}
