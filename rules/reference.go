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

func (this sourceReference) Namespace() string {
	return this.namespace
}

func (this sourceReference) Name() string {
	return this.name
}

func (this sourceReference) Type() string {
	return "ingress"
}

func (this sourceReference) String() string {
	return this.Type() + ":" + this.Namespace() + "/" + this.Name()
}

func (this sourceReference) Equals(o SourceReference) bool {
	if o == nil {
		return false
	} else {
		return o.Type() == this.Type() &&
			o.Namespace() == this.Namespace() &&
			o.Name() == this.Name()
	}
}
