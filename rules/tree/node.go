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

func (this *node[T]) hasContent() bool {
	if this.children != nil && len(this.children) > 0 {
		return true
	}
	if this.elements != nil && len(this.elements) > 0 {
		return true
	}
	return false
}

func (this *node[T]) clone() *node[T] {
	var children map[string]*node[T]
	if this.children != nil {
		children = make(map[string]*node[T])
		for key, child := range this.children {
			children[key] = child.clone()
		}
	}

	var elements map[string][]T
	if this.elements != nil {
		elements = make(map[string][]T)
		for key, sourceElements := range this.elements {
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
