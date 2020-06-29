package rules

type SourceReference interface {
	Namespace() string
	Name() string
	Type() string
	String() string
	Equals(SourceReference) bool
}

type sourceReference struct {
	namespace string
	name      string
}

func (instance sourceReference) Namespace() string {
	return instance.namespace
}

func (instance sourceReference) Name() string {
	return instance.name
}

func (instance sourceReference) Type() string {
	return "ingress"
}

func (instance sourceReference) String() string {
	return instance.Type() + ":" + instance.Namespace() + "/" + instance.Name()
}

func (instance sourceReference) Equals(o SourceReference) bool {
	if o == nil {
		return false
	} else {
		return o.Type() == instance.Type() &&
			o.Namespace() == instance.Namespace() &&
			o.Name() == instance.Name()
	}
}
