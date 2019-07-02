package tree

import (
	. "github.com/onsi/gomega"
	"reflect"
	"strings"
	"testing"
)

var (
	givenTreeForRemove = Tree{
		root: &node{
			children: map[string]*node{
				"a1": {
					children: map[string]*node{
						"b1": {
							elements: map[string][]interface{}{
								"c1": {"elementA1B1C1!", "elementA1B1C1?", "elementA1B1C1-"},
								"c2": {"elementA1B1C2!", "elementA1B1C2?", "elementA1B1C2-"},
								"c3": {"elementA1B1C3!", "elementA1B1C3?", "elementA1B1C3-"},
							},
						},
						"b2": {
							elements: map[string][]interface{}{
								"c1": {"elementA1B2C1!", "elementA1B2C1?", "elementA1B2C1-"},
								"c2": {"elementA1B2C2!", "elementA1B2C2?", "elementA1B2C2-"},
								"c3": {"elementA1B2C3!", "elementA1B2C3?", "elementA1B2C3-"},
							},
						},
					},
					elements: map[string][]interface{}{
						"b1": {"elementA1B1!", "elementA1B1?", "elementA1B2-"},
						"b2": {"elementA1B2!", "elementA1B2?", "elementA1B2-"},
						"b3": {"elementA1B3!", "elementA1B3?", "elementA1B3-"},
					},
				},
				"a2": {
					children: map[string]*node{
						"b1": {
							elements: map[string][]interface{}{
								"c1": {"elementA2B1C1!", "elementA2B1C1?", "elementA2B1C1-"},
								"c2": {"elementA2B1C2!", "elementA2B1C2?", "elementA2B1C2-"},
								"c3": {"elementA2B1C3!", "elementA2B1C3?", "elementA2B1C3-"},
							},
						},
						"b2": {
							elements: map[string][]interface{}{
								"c1": {"elementA2B2C1!", "elementA2B2C1?", "elementA2B2C1-"},
								"c2": {"elementA2B2C2!", "elementA2B2C2?", "elementA2B2C2-"},
								"c3": {"elementA2B2C3!", "elementA2B2C3?", "elementA2B2C3-"},
							},
						},
					},
					elements: map[string][]interface{}{
						"b1": {"elementA2B1!", "elementA2B1?", "elementA2B2-"},
						"b2": {"elementA2B2!", "elementA2B2?", "elementA2B2-"},
						"b3": {"elementA2B3!", "elementA2B3?", "elementA2B3-"},
					},
				},
			},
			elements: map[string][]interface{}{
				"a1": {"elementA1!", "elementA1?", "elementA1-"},
				"a2": {"elementA2!", "elementA2?", "elementA2-"},
				"a3": {"elementA3!", "elementA3?", "elementA3-"},
			},
		},
		rootElements: &[]interface{}{"element!", "element?", "element-"},
	}
)

func Test_Node_Remove_removes_elements_from_start_of_groups_across_tree(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return strings.HasSuffix(element.(string), "!")
	}, node{
		children: map[string]*node{
			"a1": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B1C1?", "elementA1B1C1-"},
							"c2": {"elementA1B1C2?", "elementA1B1C2-"},
							"c3": {"elementA1B1C3?", "elementA1B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B2C1?", "elementA1B2C1-"},
							"c2": {"elementA1B2C2?", "elementA1B2C2-"},
							"c3": {"elementA1B2C3?", "elementA1B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA1B1?", "elementA1B2-"},
					"b2": {"elementA1B2?", "elementA1B2-"},
					"b3": {"elementA1B3?", "elementA1B3-"},
				},
			},
			"a2": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B1C1?", "elementA2B1C1-"},
							"c2": {"elementA2B1C2?", "elementA2B1C2-"},
							"c3": {"elementA2B1C3?", "elementA2B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B2C1?", "elementA2B2C1-"},
							"c2": {"elementA2B2C2?", "elementA2B2C2-"},
							"c3": {"elementA2B2C3?", "elementA2B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA2B1?", "elementA2B2-"},
					"b2": {"elementA2B2?", "elementA2B2-"},
					"b3": {"elementA2B3?", "elementA2B3-"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a1": {"elementA1?", "elementA1-"},
			"a2": {"elementA2?", "elementA2-"},
			"a3": {"elementA3?", "elementA3-"},
		},
	}, &[]interface{}{"element?", "element-"})
}

func Test_Node_Remove_removes_elements_from_end_of_groups_across_tree(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return strings.HasSuffix(element.(string), "-")
	}, node{
		children: map[string]*node{
			"a1": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B1C1!", "elementA1B1C1?"},
							"c2": {"elementA1B1C2!", "elementA1B1C2?"},
							"c3": {"elementA1B1C3!", "elementA1B1C3?"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B2C1!", "elementA1B2C1?"},
							"c2": {"elementA1B2C2!", "elementA1B2C2?"},
							"c3": {"elementA1B2C3!", "elementA1B2C3?"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA1B1!", "elementA1B1?"},
					"b2": {"elementA1B2!", "elementA1B2?"},
					"b3": {"elementA1B3!", "elementA1B3?"},
				},
			},
			"a2": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B1C1!", "elementA2B1C1?"},
							"c2": {"elementA2B1C2!", "elementA2B1C2?"},
							"c3": {"elementA2B1C3!", "elementA2B1C3?"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B2C1!", "elementA2B2C1?"},
							"c2": {"elementA2B2C2!", "elementA2B2C2?"},
							"c3": {"elementA2B2C3!", "elementA2B2C3?"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA2B1!", "elementA2B1?"},
					"b2": {"elementA2B2!", "elementA2B2?"},
					"b3": {"elementA2B3!", "elementA2B3?"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a1": {"elementA1!", "elementA1?"},
			"a2": {"elementA2!", "elementA2?"},
			"a3": {"elementA3!", "elementA3?"},
		},
	}, &[]interface{}{"element!", "element?"})
}

func Test_Node_Remove_removes_elements_from_center_of_groups_across_tree(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return strings.HasSuffix(element.(string), "?")
	}, node{
		children: map[string]*node{
			"a1": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B1C1!", "elementA1B1C1-"},
							"c2": {"elementA1B1C2!", "elementA1B1C2-"},
							"c3": {"elementA1B1C3!", "elementA1B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B2C1!", "elementA1B2C1-"},
							"c2": {"elementA1B2C2!", "elementA1B2C2-"},
							"c3": {"elementA1B2C3!", "elementA1B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA1B1!", "elementA1B2-"},
					"b2": {"elementA1B2!", "elementA1B2-"},
					"b3": {"elementA1B3!", "elementA1B3-"},
				},
			},
			"a2": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B1C1!", "elementA2B1C1-"},
							"c2": {"elementA2B1C2!", "elementA2B1C2-"},
							"c3": {"elementA2B1C3!", "elementA2B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B2C1!", "elementA2B2C1-"},
							"c2": {"elementA2B2C2!", "elementA2B2C2-"},
							"c3": {"elementA2B2C3!", "elementA2B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA2B1!", "elementA2B2-"},
					"b2": {"elementA2B2!", "elementA2B2-"},
					"b3": {"elementA2B3!", "elementA2B3-"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a1": {"elementA1!", "elementA1-"},
			"a2": {"elementA2!", "elementA2-"},
			"a3": {"elementA3!", "elementA3-"},
		},
	}, &[]interface{}{"element!", "element-"})
}

func Test_Node_Remove_removes_elements_from_beginning_of_tree(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return reflect.DeepEqual(path, []string{"a1"}) || reflect.DeepEqual(path, []string{"a1", "b1"}) || reflect.DeepEqual(path, []string{"a1", "b1", "c1"})
	}, node{
		children: map[string]*node{
			"a1": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c2": {"elementA1B1C2!", "elementA1B1C2?", "elementA1B1C2-"},
							"c3": {"elementA1B1C3!", "elementA1B1C3?", "elementA1B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B2C1!", "elementA1B2C1?", "elementA1B2C1-"},
							"c2": {"elementA1B2C2!", "elementA1B2C2?", "elementA1B2C2-"},
							"c3": {"elementA1B2C3!", "elementA1B2C3?", "elementA1B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b2": {"elementA1B2!", "elementA1B2?", "elementA1B2-"},
					"b3": {"elementA1B3!", "elementA1B3?", "elementA1B3-"},
				},
			},
			"a2": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B1C1!", "elementA2B1C1?", "elementA2B1C1-"},
							"c2": {"elementA2B1C2!", "elementA2B1C2?", "elementA2B1C2-"},
							"c3": {"elementA2B1C3!", "elementA2B1C3?", "elementA2B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B2C1!", "elementA2B2C1?", "elementA2B2C1-"},
							"c2": {"elementA2B2C2!", "elementA2B2C2?", "elementA2B2C2-"},
							"c3": {"elementA2B2C3!", "elementA2B2C3?", "elementA2B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA2B1!", "elementA2B1?", "elementA2B2-"},
					"b2": {"elementA2B2!", "elementA2B2?", "elementA2B2-"},
					"b3": {"elementA2B3!", "elementA2B3?", "elementA2B3-"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a2": {"elementA2!", "elementA2?", "elementA2-"},
			"a3": {"elementA3!", "elementA3?", "elementA3-"},
		},
	}, &[]interface{}{"element!", "element?", "element-"})
}

func Test_Node_Remove_removes_elements_from_end_of_tree(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return reflect.DeepEqual(path, []string{"a3"}) || reflect.DeepEqual(path, []string{"a1", "b3"}) || reflect.DeepEqual(path, []string{"a1", "b1", "c3"})
	}, node{
		children: map[string]*node{
			"a1": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B1C1!", "elementA1B1C1?", "elementA1B1C1-"},
							"c2": {"elementA1B1C2!", "elementA1B1C2?", "elementA1B1C2-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B2C1!", "elementA1B2C1?", "elementA1B2C1-"},
							"c2": {"elementA1B2C2!", "elementA1B2C2?", "elementA1B2C2-"},
							"c3": {"elementA1B2C3!", "elementA1B2C3?", "elementA1B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA1B1!", "elementA1B1?", "elementA1B2-"},
					"b2": {"elementA1B2!", "elementA1B2?", "elementA1B2-"},
				},
			},
			"a2": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B1C1!", "elementA2B1C1?", "elementA2B1C1-"},
							"c2": {"elementA2B1C2!", "elementA2B1C2?", "elementA2B1C2-"},
							"c3": {"elementA2B1C3!", "elementA2B1C3?", "elementA2B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B2C1!", "elementA2B2C1?", "elementA2B2C1-"},
							"c2": {"elementA2B2C2!", "elementA2B2C2?", "elementA2B2C2-"},
							"c3": {"elementA2B2C3!", "elementA2B2C3?", "elementA2B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA2B1!", "elementA2B1?", "elementA2B2-"},
					"b2": {"elementA2B2!", "elementA2B2?", "elementA2B2-"},
					"b3": {"elementA2B3!", "elementA2B3?", "elementA2B3-"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a1": {"elementA1!", "elementA1?", "elementA1-"},
			"a2": {"elementA2!", "elementA2?", "elementA2-"},
		},
	}, &[]interface{}{"element!", "element?", "element-"})
}

func Test_Node_Remove_removes_elements_from_center_of_tree(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return reflect.DeepEqual(path, []string{"a2"}) || reflect.DeepEqual(path, []string{"a1", "b2"}) || reflect.DeepEqual(path, []string{"a1", "b1", "c2"})
	}, node{
		children: map[string]*node{
			"a1": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B1C1!", "elementA1B1C1?", "elementA1B1C1-"},
							"c3": {"elementA1B1C3!", "elementA1B1C3?", "elementA1B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B2C1!", "elementA1B2C1?", "elementA1B2C1-"},
							"c2": {"elementA1B2C2!", "elementA1B2C2?", "elementA1B2C2-"},
							"c3": {"elementA1B2C3!", "elementA1B2C3?", "elementA1B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA1B1!", "elementA1B1?", "elementA1B2-"},
					"b3": {"elementA1B3!", "elementA1B3?", "elementA1B3-"},
				},
			},
			"a2": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B1C1!", "elementA2B1C1?", "elementA2B1C1-"},
							"c2": {"elementA2B1C2!", "elementA2B1C2?", "elementA2B1C2-"},
							"c3": {"elementA2B1C3!", "elementA2B1C3?", "elementA2B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B2C1!", "elementA2B2C1?", "elementA2B2C1-"},
							"c2": {"elementA2B2C2!", "elementA2B2C2?", "elementA2B2C2-"},
							"c3": {"elementA2B2C3!", "elementA2B2C3?", "elementA2B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA2B1!", "elementA2B1?", "elementA2B2-"},
					"b2": {"elementA2B2!", "elementA2B2?", "elementA2B2-"},
					"b3": {"elementA2B3!", "elementA2B3?", "elementA2B3-"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a1": {"elementA1!", "elementA1?", "elementA1-"},
			"a3": {"elementA3!", "elementA3?", "elementA3-"},
		},
	}, &[]interface{}{"element!", "element?", "element-"})
}

func Test_Node_Remove_removes_whole_branch_at_end_of_tree(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return len(path) == 3 &&
			(path[0] == "a1" || path[0] == "a2") &&
			path[1] == "b1"
	}, node{
		children: map[string]*node{
			"a1": {
				children: map[string]*node{
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B2C1!", "elementA1B2C1?", "elementA1B2C1-"},
							"c2": {"elementA1B2C2!", "elementA1B2C2?", "elementA1B2C2-"},
							"c3": {"elementA1B2C3!", "elementA1B2C3?", "elementA1B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA1B1!", "elementA1B1?", "elementA1B2-"},
					"b2": {"elementA1B2!", "elementA1B2?", "elementA1B2-"},
					"b3": {"elementA1B3!", "elementA1B3?", "elementA1B3-"},
				},
			},
			"a2": {
				children: map[string]*node{
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B2C1!", "elementA2B2C1?", "elementA2B2C1-"},
							"c2": {"elementA2B2C2!", "elementA2B2C2?", "elementA2B2C2-"},
							"c3": {"elementA2B2C3!", "elementA2B2C3?", "elementA2B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA2B1!", "elementA2B1?", "elementA2B2-"},
					"b2": {"elementA2B2!", "elementA2B2?", "elementA2B2-"},
					"b3": {"elementA2B3!", "elementA2B3?", "elementA2B3-"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a1": {"elementA1!", "elementA1?", "elementA1-"},
			"a2": {"elementA2!", "elementA2?", "elementA2-"},
			"a3": {"elementA3!", "elementA3?", "elementA3-"},
		},
	}, &[]interface{}{"element!", "element?", "element-"})
}

func Test_Node_Remove_removes_whole_branch_at_center_of_tree(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return len(path) >= 1 &&
			path[0] == "a1"
	}, node{
		children: map[string]*node{
			"a2": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B1C1!", "elementA2B1C1?", "elementA2B1C1-"},
							"c2": {"elementA2B1C2!", "elementA2B1C2?", "elementA2B1C2-"},
							"c3": {"elementA2B1C3!", "elementA2B1C3?", "elementA2B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B2C1!", "elementA2B2C1?", "elementA2B2C1-"},
							"c2": {"elementA2B2C2!", "elementA2B2C2?", "elementA2B2C2-"},
							"c3": {"elementA2B2C3!", "elementA2B2C3?", "elementA2B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA2B1!", "elementA2B1?", "elementA2B2-"},
					"b2": {"elementA2B2!", "elementA2B2?", "elementA2B2-"},
					"b3": {"elementA2B3!", "elementA2B3?", "elementA2B3-"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a2": {"elementA2!", "elementA2?", "elementA2-"},
			"a3": {"elementA3!", "elementA3?", "elementA3-"},
		},
	}, &[]interface{}{"element!", "element?", "element-"})
}

func Test_Node_Remove_removes_just_rootElements(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return len(path) == 0
	}, node{
		children: map[string]*node{
			"a1": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B1C1!", "elementA1B1C1?", "elementA1B1C1-"},
							"c2": {"elementA1B1C2!", "elementA1B1C2?", "elementA1B1C2-"},
							"c3": {"elementA1B1C3!", "elementA1B1C3?", "elementA1B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA1B2C1!", "elementA1B2C1?", "elementA1B2C1-"},
							"c2": {"elementA1B2C2!", "elementA1B2C2?", "elementA1B2C2-"},
							"c3": {"elementA1B2C3!", "elementA1B2C3?", "elementA1B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA1B1!", "elementA1B1?", "elementA1B2-"},
					"b2": {"elementA1B2!", "elementA1B2?", "elementA1B2-"},
					"b3": {"elementA1B3!", "elementA1B3?", "elementA1B3-"},
				},
			},
			"a2": {
				children: map[string]*node{
					"b1": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B1C1!", "elementA2B1C1?", "elementA2B1C1-"},
							"c2": {"elementA2B1C2!", "elementA2B1C2?", "elementA2B1C2-"},
							"c3": {"elementA2B1C3!", "elementA2B1C3?", "elementA2B1C3-"},
						},
					},
					"b2": {
						elements: map[string][]interface{}{
							"c1": {"elementA2B2C1!", "elementA2B2C1?", "elementA2B2C1-"},
							"c2": {"elementA2B2C2!", "elementA2B2C2?", "elementA2B2C2-"},
							"c3": {"elementA2B2C3!", "elementA2B2C3?", "elementA2B2C3-"},
						},
					},
				},
				elements: map[string][]interface{}{
					"b1": {"elementA2B1!", "elementA2B1?", "elementA2B2-"},
					"b2": {"elementA2B2!", "elementA2B2?", "elementA2B2-"},
					"b3": {"elementA2B3!", "elementA2B3?", "elementA2B3-"},
				},
			},
		},
		elements: map[string][]interface{}{
			"a1": {"elementA1!", "elementA1?", "elementA1-"},
			"a2": {"elementA2!", "elementA2?", "elementA2-"},
			"a3": {"elementA3!", "elementA3?", "elementA3-"},
		},
	}, nil)
}

func Test_Node_Remove_removes_whole_tree(t *testing.T) {
	executeTestRemoveRun(t, func(path []string, element interface{}) bool {
		return true
	}, node{}, nil)
}

func executeTestRemoveRun(t *testing.T, predicate Predicate, expected node, expectedRootElements *[]interface{}) {
	g := NewGomegaWithT(t)

	instance := givenTreeForRemove.Clone()

	err := instance.Remove(predicate)

	g.Expect(err).To(BeNil())
	g.Expect(instance.root).To(Equal(&expected))
	g.Expect(instance.rootElements).To(Equal(expectedRootElements))
}
