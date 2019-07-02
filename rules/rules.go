package rules

import "fmt"

type Rules interface {
	Get(int) Rule
	Len() int
	Any() Rule
}

type rules []interface{}

func (instance rules) Get(i int) Rule {
	return instance[i].(Rule)
}

func (instance rules) Len() int {
	if instance == nil {
		return 0
	}
	return len(instance)
}

func (instance rules) Any() Rule {
	if len(instance) > 0 {
		return (instance)[0].(Rule)
	}
	return nil
}

func (instance rules) String() string {
	result := ""
	for i, r := range instance {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprint(r)
	}
	return result
}
