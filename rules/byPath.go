package rules

import "github.com/echocat/lingress/rules/tree"

type ByPath struct {
	values *tree.Tree[Rule]

	onAdded   OnAdded
	onRemoved OnRemoved
}

func NewByPath(onAdded OnAdded, onRemoved OnRemoved) *ByPath {
	result := &ByPath{
		values:    tree.New[Rule](),
		onAdded:   onAdded,
		onRemoved: onRemoved,
	}

	result.values.OnAdded = func(path []string, r Rule) {
		onAdded(path, r)
	}
	result.values.OnRemoved = func(path []string, r Rule) {
		onRemoved(path, r)
	}
	return result
}

func (instance *ByPath) All(consumer func(Rule) error) error {
	return instance.values.All(consumer)
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
	return instance.values.Remove(func(path []string, r Rule) bool {
		return predicate(path, r)
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
