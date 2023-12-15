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

func (this *ByPath) All(consumer func(Rule) error) error {
	return this.values.All(consumer)
}

func (this *ByPath) Find(path []string) (Rules, error) {
	if plain, err := this.values.Find(path); err != nil {
		return nil, err
	} else {
		return rules(plain), err
	}
}

func (this *ByPath) Put(r Rule) error {
	return this.values.Put(r.Path(), r)
}

func (this *ByPath) Remove(predicate Predicate) error {
	return this.values.Remove(func(path []string, r Rule) bool {
		return predicate(path, r)
	})
}

func (this *ByPath) HasContent() bool {
	return this.values.HasContent()
}

func (this *ByPath) Clone() *ByPath {
	result := NewByPath(this.onAdded, this.onRemoved)
	result.values = this.values.Clone()
	return result
}
