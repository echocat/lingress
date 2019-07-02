package rules

type SourceReference interface {
	Name() string
	Type() string
	String() string
	Equals(SourceReference) bool
}

type sourceReference struct {
	namespace string
	name      string
}

func (instance sourceReference) Name() string {
	return instance.namespace + "/" + instance.name
}

func (instance sourceReference) Type() string {
	return "ingress"
}

func (instance sourceReference) String() string {
	return instance.Type() + ":" + instance.Name()
}

func (instance sourceReference) Equals(o SourceReference) bool {
	if o == nil {
		return false
	} else if sr, ok := o.(sourceReference); ok {
		return sr.namespace == instance.namespace &&
			sr.name == instance.name
	} else if sr, ok := o.(*sourceReference); ok {
		return sr.namespace == instance.namespace &&
			sr.name == instance.name
	} else {
		return o.Name() == instance.Name()
	}
}
