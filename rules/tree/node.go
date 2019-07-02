package tree

type node struct {
	children map[string]*node
	elements map[string][]interface{}
}

func newNode() *node {
	return &node{}
}

type Cloneable interface {
	Clone() interface{}
}

func (instance *node) hasContent() bool {
	if instance.children != nil && len(instance.children) > 0 {
		return true
	}
	if instance.elements != nil && len(instance.elements) > 0 {
		return true
	}
	return false
}

func (instance *node) clone() *node {
	var children map[string]*node
	if instance.children != nil {
		children = make(map[string]*node)
		for key, child := range instance.children {
			children[key] = child.clone()
		}
	}

	var elements map[string][]interface{}
	if instance.elements != nil {
		elements = make(map[string][]interface{})
		for key, sourceElements := range instance.elements {
			elements[key] = cloneElements(sourceElements)
		}
	}

	return &node{
		children: children,
		elements: elements,
	}
}

func cloneElements(in []interface{}) (out []interface{}) {
	out = make([]interface{}, len(in))
	for i, sourceElement := range in {
		if cloneable, ok := sourceElement.(Cloneable); ok {
			out[i] = cloneable.Clone()
		} else {
			out[i] = sourceElement
		}
	}
	return out
}
