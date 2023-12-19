package tree

import (
	. "github.com/onsi/gomega"
	"strings"
	"testing"
)

var (
	givenTreeForFind = Tree[testValue]{
		root: &node[testValue]{
			children: map[string]*node[testValue]{
				"a1": {
					children: map[string]*node[testValue]{
						"b1": {
							elements: map[string][]testValue{
								"c1": {"A1B1C1"},
								"c2": {"A1B1C2"},
								"c3": {"A1B1C3"},
							},
						},
						"b2": {
							elements: map[string][]testValue{
								"c1": {"A1B2C1"},
								"c2": {"A1B2C2"},
								"c3": {"A1B2C3"},
							},
						},
					},
					elements: map[string][]testValue{
						"b1": {"A1B1"},
						"b3": {"A1B3"},
					},
				},
				"a2": {
					children: map[string]*node[testValue]{
						"b1": {
							elements: map[string][]testValue{
								"c1": {"A2B1C1"},
								"c2": {"A2B1C2"},
								"c3": {"A2B1C3"},
							},
						},
						"b2": {
							elements: map[string][]testValue{
								"c1": {"A2B2C1"},
								"c2": {"A2B2C2"},
								"c3": {"A2B2C3"},
							},
						},
						"b3": {
							children: map[string]*node[testValue]{
								"c1": {
									elements: map[string][]testValue{
										"d1": {"A2B3C1"},
										"d2": {"A2B3C1"},
									},
								},
							},
						},
					},
					elements: map[string][]testValue{
						"b1": {"A2B1"},
						"b2": {"A2B2"},
						"b3": {"A2B3"},
					},
				},
			},
			elements: map[string][]testValue{
				"a1": {"A1"},
				"a2": {"A2"},
				"a3": {"A3"},
			},
		},
		rootElements: &[]testValue{"ROOT"},
	}
)

func Test_Node_Find(t *testing.T) {
	executeTestFindRun(t, "/", "ROOT")
	executeTestFindRun(t, "/a1", "A1")
	executeTestFindRun(t, "/a1/b1", "A1B1")
	executeTestFindRun(t, "/a1/b1/c1", "A1B1C1")
	executeTestFindRun(t, "/a1/b2", "A1")
	executeTestFindRun(t, "/a2/b2", "A2B2")
	executeTestFindRun(t, "/a2/b2/c4", "A2B2")
	executeTestFindRun(t, "/a3/b2/c4", "A3")
	executeTestFindRun(t, "/a1/b1/c1/x1", "A1B1C1")
	executeTestFindRun(t, "/a2", "A2")
	executeTestFindRun(t, "/xxx/a2", "ROOT")
}

func executeTestFindRun(t *testing.T, path string, expectedElements ...testValue) {
	name := strings.ReplaceAll(path[1:], "/", "_")
	if path == "/" {
		name = "ROOT"
	}
	t.Run(name, func(t *testing.T) {
		g := NewGomegaWithT(t)

		instance := givenTreeForFind

		pathElements := strings.Split(path, "/")
		actual, err := instance.Find(pathElements[1:])

		g.Expect(err).To(BeNil())
		g.Expect(actual).To(Equal(expectedElements))
	})
}
