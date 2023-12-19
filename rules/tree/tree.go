package tree

type Type[T any] interface {
	Cloneable[T]
}

type Tree[T Type[T]] struct {
	root         *node[T]
	rootElements *[]T

	OnAdded   OnAdded[T]
	OnRemoved OnRemoved[T]
}

type OnAdded[T Type[T]] func(path []string, element T)
type OnRemoved[T Type[T]] func(path []string, element T)

func New[T Type[T]]() *Tree[T] {
	return &Tree[T]{
		root: newNode[T](),
	}
}

type Predicate[T Type[T]] func(path []string, element T) bool

func (this *Tree[T]) All(consumer func(T) error) error {
	if n := this.root; n != nil {
		if err := this.all(consumer, n); err != nil {
			return err
		}
	}
	if es := this.rootElements; es != nil {
		for _, e := range *es {
			if err := consumer(e); err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *Tree[T]) all(consumer func(T) error, n *node[T]) error {
	if cs := n.children; cs != nil {
		for _, c := range cs {
			if err := this.all(consumer, c); err != nil {
				return err
			}
		}
	}
	if n2es := n.elements; n2es != nil {
		for _, es := range n2es {
			for _, e := range es {
				if err := consumer(e); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (this *Tree[T]) Put(path []string, element T) error {
	this.put(path, element)
	return nil
}

func (this *Tree[T]) Remove(predicate Predicate[T]) error {
	this.remove([]string{}, this.root, predicate)
	if this.rootElements != nil {
		buf := this.removeElement(*this.rootElements, []string{}, predicate)
		if len(buf) > 0 {
			this.rootElements = &buf
		} else {
			this.rootElements = nil
		}
	}
	return nil
}

func (this *Tree[T]) Find(path []string) (result []T, err error) {
	return this.find(path), nil
}

func (this *Tree[T]) Clone() *Tree[T] {
	var rootElements *[]T
	if this.rootElements != nil {
		buf := cloneElements(*this.rootElements)
		rootElements = &buf
	}
	return &Tree[T]{
		root:         this.root.clone(),
		rootElements: rootElements,
	}
}

func (this *Tree[T]) HasContent() bool {
	if this.root != nil && this.root.hasContent() {
		return true
	}

	if this.rootElements != nil && len(*this.rootElements) > 0 {
		return true
	}

	return false
}

func (this *Tree[T]) put(path []string, element T) {
	if len(path) == 0 {
		if this.rootElements == nil {
			this.rootElements = &[]T{element}
		} else {
			buf := append(*this.rootElements, element)
			this.rootElements = &buf
		}
		return
	}

	current := this.root
	for i := 0; i < len(path); i++ {
		key := path[i]
		if i+1 < len(path) {
			childSelected := false
			if current.children != nil {
				if child, ok := current.children[key]; ok {
					current = child
					childSelected = true
				}
			}
			if !childSelected {
				newCurrent := newNode[T]()
				if current.children == nil {
					current.children = make(map[string]*node[T])
				}
				current.children[key] = newCurrent
				current = newCurrent
			}
		} else {
			if current.elements == nil {
				current.elements = make(map[string][]T)
			}
			if existing, ok := current.elements[key]; ok {
				current.elements[key] = append(existing, element)
			} else {
				current.elements[key] = []T{element}
			}

			el := this.OnAdded
			if el != nil {
				el(path, element)
			}
		}
	}
}

func (this *Tree[T]) remove(path []string, n *node[T], predicate Predicate[T]) {
	if n.children != nil {
		for key, child := range n.children {
			childPath := append(path, key)
			this.remove(childPath, child, predicate)
			if !child.hasContent() {
				delete(n.children, key)
			}
		}
		if len(n.children) <= 0 {
			n.children = nil
		}
	}
	if n.elements != nil {
		for key, elements := range n.elements {
			elementPath := append(path, key)
			elements = this.removeElement(elements, elementPath, predicate)
			if len(elements) == 0 {
				delete(n.elements, key)
			} else {
				n.elements[key] = elements
			}
		}
		if len(n.elements) <= 0 {
			n.elements = nil
		}
	}
}

func (this *Tree[T]) find(path []string) (result []T) {
	if this.rootElements != nil {
		result = *this.rootElements
	}

	current := this.root

	for _, key := range path {
		matchInPart := false
		if current.elements != nil {
			if candidate, ok := current.elements[key]; ok {
				result = candidate
			}
		}
		if current.children != nil {
			if child, ok := current.children[key]; ok {
				current = child
				matchInPart = true
			}
		}
		if !matchInPart {
			break
		}
	}

	return
}

func (this *Tree[T]) removeElement(elements []T, path []string, predicate Predicate[T]) []T {
	for next := true; next; {
		next = false

		for i, candidate := range elements {
			if predicate(path, candidate) {
				buf := make([]T, len(elements)-1)
				copy(buf, elements[:i])
				copy(buf[i:], elements[i+1:])
				elements = buf
				next = true

				el := this.OnRemoved
				if el != nil {
					el(path, candidate)
				}

				break
			}
		}
	}
	return elements
}
