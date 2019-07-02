package rules

import "github.com/echocat/lingress/rules/tree"

type ByPath struct {
	values *tree.Tree

	onAdded   OnAdded
	onRemoved OnRemoved
}

func NewByPath(onAdded OnAdded, onRemoved OnRemoved) *ByPath {
	result := &ByPath{
		values:    tree.New(),
		onAdded:   onAdded,
		onRemoved: onRemoved,
	}

	result.values.OnAdded = func(path []string, element interface{}) {
		if r, ok := element.(Rule); ok {
			onAdded(path, r)
		}
	}
	result.values.OnRemoved = func(path []string, element interface{}) {
		if r, ok := element.(Rule); ok {
			onRemoved(path, r)
		}
	}
	return result
}

func (instance *ByPath) All(consumer func(Rule) error) error {
	return instance.values.All(func(value interface{}) error {
		return consumer(value.(Rule))
	})
}

func (instance *ByPath) Find(path []string) (Rules, error) {
	if plain, err := instance.values.Find(path); err != nil {
		return nil, err
	} else {
		return rules(plain), err
	}
}

func (instance *ByPath) Put(r Rule) error {
	return instance.values.Put(r.Path(), r)
}

func (instance *ByPath) Remove(predicate Predicate) error {
	return instance.values.Remove(func(path []string, element interface{}) bool {
		if r, ok := element.(Rule); ok {
			return predicate(path, r)
		} else {
			return false
		}
	})
}

func (instance *ByPath) HasContent() bool {
	return instance.values.HasContent()
}

func (instance *ByPath) Clone() *ByPath {
	result := NewByPath(instance.onAdded, instance.onRemoved)
	result.values = instance.values.Clone()
	return result
}
