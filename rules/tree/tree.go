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

func (instance *Tree[T]) All(consumer func(T) error) error {
	if n := instance.root; n != nil {
		if err := instance.all(consumer, n); err != nil {
			return err
		}
	}
	if es := instance.rootElements; es != nil {
		for _, e := range *es {
			if err := consumer(e); err != nil {
				return err
			}
		}
	}
	return nil
}

func (instance *Tree[T]) all(consumer func(T) error, n *node[T]) error {
	if cs := n.children; cs != nil {
		for _, c := range cs {
			if err := instance.all(consumer, c); err != nil {
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

func (instance *Tree[T]) Put(path []string, element T) error {
	instance.put(path, element)
	return nil
}

func (instance *Tree[T]) Remove(predicate Predicate[T]) error {
	instance.remove([]string{}, instance.root, predicate)
	if instance.rootElements != nil {
		buf := instance.removeElement(*instance.rootElements, []string{}, predicate)
		if len(buf) > 0 {
			instance.rootElements = &buf
		} else {
			instance.rootElements = nil
		}
	}
	return nil
}

func (instance *Tree[T]) Find(path []string) (result []T, err error) {
	return instance.find(path), nil
}

func (instance *Tree[T]) Clone() *Tree[T] {
	var rootElements *[]T
	if instance.rootElements != nil {
		buf := cloneElements(*instance.rootElements)
		rootElements = &buf
	}
	return &Tree[T]{
		root:         instance.root.clone(),
		rootElements: rootElements,
	}
}

func (instance *Tree[T]) HasContent() bool {
	if instance.root != nil && instance.root.hasContent() {
		return true
	}

	if instance.rootElements != nil && len(*instance.rootElements) > 0 {
		return true
	}

	return false
}

func (instance *Tree[T]) put(path []string, element T) {
	if len(path) == 0 {
		if instance.rootElements == nil {
			instance.rootElements = &[]T{element}
		} else {
			buf := append(*instance.rootElements, element)
			instance.rootElements = &buf
		}
		return
	}

	current := instance.root
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

			el := instance.OnAdded
			if el != nil {
				el(path, element)
			}
		}
	}
}

func (instance *Tree[T]) remove(path []string, n *node[T], predicate Predicate[T]) {
	if n.children != nil {
		for key, child := range n.children {
			childPath := append(path, key)
			instance.remove(childPath, child, predicate)
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
			elements = instance.removeElement(elements, elementPath, predicate)
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

func (instance *Tree[T]) find(path []string) (result []T) {
	if instance.rootElements != nil {
		result = *instance.rootElements
	}

	current := instance.root

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

func (instance *Tree[T]) removeElement(elements []T, path []string, predicate Predicate[T]) []T {
	for next := true; next; {
		next = false

		for i, candidate := range elements {
			if predicate(path, candidate) {
				buf := make([]T, len(elements)-1)
				copy(buf, elements[:i])
				copy(buf[i:], elements[i+1:])
				elements = buf
				next = true

				el := instance.OnRemoved
				if el != nil {
					el(path, candidate)
				}

				break
			}
		}
	}
	return elements
}
