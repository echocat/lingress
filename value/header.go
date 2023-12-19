package value

import (
	"fmt"
	"net/textproto"
	"strings"
)

type Header struct {
	Key    string
	Value  string
	Forced bool
	Add    bool
	Del    bool
}

func (this *Header) Set(plain string) error {
	in := plain
	r := Header{}

	if len(in) > 0 && in[0] == '!' {
		r.Forced = true
		in = in[1:]
	}

	if len(in) > 0 && in[0] == '-' {
		r.Del = true
		in = in[1:]
	} else if len(in) > 0 && in[0] == '+' {
		r.Add = true
		in = in[1:]
	}

	ci := strings.IndexRune(in, ':')
	if ci >= 0 && r.Del {
		return fmt.Errorf("illegal header rule '%s': delete actions should not have a value", plain)
	}
	if ci < 0 && !r.Del {
		return fmt.Errorf("illegal header rule '%s': add or set actions should have a value", plain)
	}

	if ci >= 0 {
		r.Key = textproto.CanonicalMIMEHeaderKey(in[:ci])
		r.Value = strings.TrimSpace(in[ci+1:])
	} else {
		r.Key = textproto.CanonicalMIMEHeaderKey(in)
	}
	if r.Key == "" {
		return fmt.Errorf("illegal header rule '%s': key is empty", plain)
	}

	*this = r
	return nil
}

func (this Header) String() string {
	result := ""
	if this.Forced {
		result += "!"
	}
	if this.Del {
		result += "-"
	} else if this.Add {
		result += "+"
	}

	result += this.Key

	if !this.Del {
		result += ":" + this.Value
	}

	return result
}

type Headers []Header

func (this Headers) String() string {
	result := ""
	for i, val := range this {
		if i > 0 {
			result += "\n"
		}
		result += val.String()
	}
	return result
}

func (this *Headers) IsCumulative() bool {
	return true
}

func (this *Headers) Set(plain string) error {
	var v Header
	if err := v.Set(plain); err != nil {
		return err
	}
	*this = append(*this, v)
	return nil
}
