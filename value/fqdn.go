package value

import (
	"errors"
	"fmt"
	"golang.org/x/net/idna"
	"strings"
	"unicode/utf8"
)

var (
	ErrIllegalFqdn = errors.New("illegal fqdn")
)

type Fqdn string

func (instance *Fqdn) UnmarshalText(b []byte) error {
	target := strings.ToLower(string(b))
	candidate := strings.Split(target, ".")

	if err := validateFqdnSegments(candidate, string(b), false); err != nil {
		return err
	}

	if strings.HasSuffix(target, ".") && len(target) > 1 {
		target = target[:len(target)-1]
	}

	*instance = Fqdn(target)
	return nil
}

func (instance Fqdn) Parent() Fqdn {
	parts := strings.SplitN(string(instance), ".", 2)

	// Too few segments...
	if len(parts) <= 1 {
		return ""
	}

	var result Fqdn
	if result.Set(parts[1]) != nil {
		// The rest is not a valid Fqdn, skipping...
		return ""
	}

	return result
}

func (instance *Fqdn) Set(plain string) error {
	return instance.UnmarshalText([]byte(plain))
}

func (instance Fqdn) MarshalText() (text []byte, err error) {
	if err := validateFqdn(string(instance), false); err != nil {
		return nil, err
	}
	return []byte(instance.String()), nil
}

func (instance Fqdn) String() string {
	return string(instance)
}

func (instance Fqdn) Get() interface{} {
	return instance
}

func (instance Fqdn) IsPresent() bool {
	return instance != ""
}

type WildcardSupportingFqdn Fqdn

func (instance WildcardSupportingFqdn) WithoutWildcard() (hadWildcard bool, withoutWildcard Fqdn, err error) {
	parts := strings.SplitN(string(instance), ".", 2)

	if len(parts) <= 0 {
		// Empty stays empty...
		return false, "", nil
	}

	withoutWildcard = Fqdn(instance)
	// Too few segments...
	if len(parts) >= 2 && parts[0] == "*" {
		hadWildcard = true
		withoutWildcard = Fqdn(parts[1])
	}

	err = validateFqdn(string(withoutWildcard), false)

	return
}

func (instance WildcardSupportingFqdn) MarshalText() (text []byte, err error) {
	if err := validateFqdn(string(instance), true); err != nil {
		return nil, err
	}
	return []byte(instance.String()), nil
}

func (instance *WildcardSupportingFqdn) UnmarshalText(b []byte) error {
	target := strings.ToLower(string(b))
	candidate := strings.Split(target, ".")

	if err := validateFqdnSegments(candidate, string(b), true); err != nil {
		return err
	}

	if strings.HasSuffix(target, ".") && len(target) > 1 {
		target = target[:len(target)-1]
	}

	*instance = WildcardSupportingFqdn(target)
	return nil
}

func (instance *WildcardSupportingFqdn) Set(plain string) error {
	return instance.UnmarshalText([]byte(plain))
}

func (instance WildcardSupportingFqdn) String() string {
	return string(instance)
}

func (instance WildcardSupportingFqdn) Get() interface{} {
	return instance
}

func (instance WildcardSupportingFqdn) IsPresent() bool {
	return instance != ""
}

func validateFqdn(candidate string, leadingWildcardAllowed bool) error {
	segments := strings.Split(strings.ToLower(candidate), ".")
	return validateFqdnSegments(segments, candidate, leadingWildcardAllowed)
}

func validateFqdnSegments(segments []string, original string, leadingWildcardAllowed bool) error {
	if len(segments) == 0 {
		return nil
	}
	if len(segments) == 1 && segments[0] == "" {
		return nil
	}

	var totalLength, respectSegments int

	for i, segment := range segments {
		if i > 0 && i == len(segments)-1 && segment == "" {
			// Tailing dot: ignore it.
			continue
		}
		sL, sErr := validateFqdnSegment(segment, i == 0 && leadingWildcardAllowed)
		if sErr != nil {
			return fmt.Errorf("%w '%s': %v", ErrIllegalFqdn, original, sErr)
		}
		totalLength += sL
		respectSegments = i
	}

	// Append the number of dots
	totalLength += len(segments) - 1

	if totalLength > 255 {
		return fmt.Errorf("%w '%s': too long (more than 255)", ErrIllegalFqdn, original)
	}

	if respectSegments == 1 && segments[0] == "*" {
		return fmt.Errorf("%w '%s': wildcard only fqdn is not allowed", ErrIllegalFqdn, original)
	}

	return nil
}

func validateFqdnSegment(candidate string, wildcardAllowed bool) (int, error) {
	if wildcardAllowed && candidate == "*" {
		return 1, nil
	}
	if decoded, err := (&idna.Profile{}).ToASCII(candidate); err == nil {
		candidate = decoded
	}
	var length int
	for i, w := 0, 0; i < len(candidate); i += w {
		ch, width := utf8.DecodeRuneInString(candidate[i:])
		if i == 0 {
			if !isValidFqdnCharacter(ch) {
				return 0, fmt.Errorf("segment '%s' does not start with a [a-z0-9]", candidate)
			}
		} else {
			if !isValidFqdnCharacter(ch) && ch != '-' {
				return 0, fmt.Errorf("segment '%s' contains illegal characters [a-z0-9-]", candidate)
			}
		}
		w = width
		length++
	}
	if length < 1 {
		return 0, fmt.Errorf("segment '%s' too short (less than 1)", candidate)
	}
	if length > 63 {
		return 0, fmt.Errorf("segment '%s' too long (more than 63)", candidate)
	}
	return length, nil
}

func isValidFqdnCharacter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= '0' && ch <= '9')
}
