package rules

type ByHost struct {
	values   map[string]*ByPath
	fallback *ByPath

	onAdded   OnAdded
	onRemoved OnRemoved
}

func NewByHost(onAdded OnAdded, onRemoved OnRemoved) *ByHost {
	return &ByHost{
		values:   make(map[string]*ByPath),
		fallback: NewByPath(onAdded, onRemoved),

		onAdded:   onAdded,
		onRemoved: onRemoved,
	}
}

func (instance *ByHost) All(consumer func(Rule) error) error {
	for _, value := range instance.values {
		if err := value.All(consumer); err != nil {
			return err
		}
	}
	if err := instance.fallback.All(consumer); err != nil {
		return err
	}
	return nil
}

func (instance *ByHost) Find(host string, path []string) (Rules, error) {
	var byHost, fallback Rules
	if v, ok := instance.values[host]; ok {
		if r, err := v.Find(path); err != nil {
			return nil, err
		} else if r != nil && r.Len() > 0 {
			byHost = r
		}
	}
	if v := instance.fallback; v != nil {
		if r, err := instance.fallback.Find(path); err != nil {
			return nil, err
		} else if r != nil && r.Len() > 0 {
			fallback = r
		}
	}

	if byHost != nil && fallback != nil {
		byHostPath := byHost.Any().Path()
		fallbackPath := fallback.Any().Path()
		if len(fallbackPath) > len(byHostPath) {
			return fallback, nil
		}
		return byHost, nil
	} else if byHost != nil {
		return byHost, nil
	} else if fallback != nil {
		return fallback, nil
	}

	return rules{}, nil
}

func (instance *ByHost) Put(r Rule) error {
	host := r.Host()

	if host == "" {
		return instance.fallback.Put(r)
	} else if existing, ok := instance.values[host]; ok {
		return existing.Put(r)
	} else {
		value := NewByPath(instance.onAdded, instance.onRemoved)
		instance.values[host] = value
		return value.Put(r)
	}
}

func (instance *ByHost) Remove(predicate Predicate) error {
	if err := instance.fallback.Remove(predicate); err != nil {
		return err
	}
	for host, v := range instance.values {
		if err := v.Remove(predicate); err != nil {
			return err
		}
		if !v.HasContent() {
			delete(instance.values, host)
		}
	}
	return nil
}

func (instance *ByHost) HasContent() bool {
	return len(instance.values) > 0
}

func (instance *ByHost) Clone() *ByHost {
	result := NewByHost(instance.onAdded, instance.onRemoved)

	result.fallback = instance.fallback.Clone()
	for k, v := range instance.values {
		result.values[k] = v.Clone()
	}

	return result
}
