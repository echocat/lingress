package tree

import (
	. "github.com/onsi/gomega"
	"strings"
	"testing"
)

func Test_Node_Put_works_with_element_in_root(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "a", element: "elementA"},
	}, node[testValue]{
		elements: map[string][]testValue{
			"a": {"elementA"},
		},
	}, nil)
}

func Test_Node_Put_works_with_element_in_subPath(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "a/b", element: "elementA"},
	}, node[testValue]{
		children: map[string]*node[testValue]{
			"a": {
				elements: map[string][]testValue{
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
	}, node[testValue]{
		elements: map[string][]testValue{
			"a": {"elementA"},
			"b": {"elementB"},
			"c": {"elementC"},
		},
	}, &[]testValue{"elementRoot1", "elementRoot2", "elementRoot3"})
}

func Test_Node_Put_works_with_3_elements_across_path(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "a", element: "elementA"},
		{path: "a/b", element: "elementAB"},
		{path: "a/b/c", element: "elementABC"},
	}, node[testValue]{
		children: map[string]*node[testValue]{
			"a": {
				children: map[string]*node[testValue]{
					"b": {
						elements: map[string][]testValue{
							"c": {"elementABC"},
						},
					},
				},
				elements: map[string][]testValue{
					"b": {"elementAB"},
				},
			},
		},
		elements: map[string][]testValue{
			"a": {"elementA"},
		},
	}, nil)
}

func Test_Node_Put_works_with_2_elements_across_path(t *testing.T) {
	executeTestPutRun(t, []putTestCaseGiven{
		{path: "a", element: "elementA"},
		{path: "a/b/c", element: "elementABC"},
	}, node[testValue]{
		children: map[string]*node[testValue]{
			"a": {
				children: map[string]*node[testValue]{
					"b": {
						elements: map[string][]testValue{
							"c": {"elementABC"},
						},
					},
				},
			},
		},
		elements: map[string][]testValue{
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
	}, node[testValue]{
		children: map[string]*node[testValue]{
			"a": {
				children: map[string]*node[testValue]{
					"b": {
						elements: map[string][]testValue{
							"c": {"elementABC!", "elementABC"},
						},
					},
				},
			},
		},
		elements: map[string][]testValue{
			"a": {"elementA!", "elementA"},
		},
	}, nil)
}

func executeTestPutRun(t *testing.T, given []putTestCaseGiven, expected node[testValue], expectedRootElements *[]testValue) {
	g := NewGomegaWithT(t)

	actual := New[testValue]()
	for _, gElement := range given {
		pathElements := strings.Split(gElement.path, "/")
		if gElement.path == "" {
			pathElements = []string{}
		}
		err := actual.Put(pathElements, testValue(gElement.element))
		g.Expect(err).To(BeNil())
	}

	g.Expect(actual.root).To(Equal(&expected))
	g.Expect(actual.rootElements).To(Equal(expectedRootElements))
}

type putTestCaseGiven struct {
	path    string
	element string
}
