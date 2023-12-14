package tree

type node[T Type[T]] struct {
	children map[string]*node[T]
	elements map[string][]T
}

func newNode[T Type[T]]() *node[T] {
	return &node[T]{}
}

type Cloneable[T any] interface {
	Clone() T
}

func (instance *node[T]) hasContent() bool {
	if instance.children != nil && len(instance.children) > 0 {
		return true
	}
	if instance.elements != nil && len(instance.elements) > 0 {
		return true
	}
	return false
}

func (instance *node[T]) clone() *node[T] {
	var children map[string]*node[T]
	if instance.children != nil {
		children = make(map[string]*node[T])
		for key, child := range instance.children {
			children[key] = child.clone()
		}
	}

	var elements map[string][]T
	if instance.elements != nil {
		elements = make(map[string][]T)
		for key, sourceElements := range instance.elements {
			elements[key] = cloneElements(sourceElements)
		}
	}

	return &node[T]{
		children: children,
		elements: elements,
	}
}

func cloneElements[T Cloneable[T]](in []T) (out []T) {
	out = make([]T, len(in))
	for i, sourceElement := range in {
		out[i] = sourceElement.Clone()
	}
	return out
}
