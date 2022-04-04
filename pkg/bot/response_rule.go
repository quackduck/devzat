package bot

import (
	"devzat/pkg/interfaces"
	"sort"
)

type responseRule struct {
	require
	with
	chance    float64
	responses []string
}

type require struct {
	exact     []string
	substring []string
}

type with struct {
	exactlyOneOf []string
	anyOneOf     []string
	allOf        []string
	noneOf       []string
}

func (r *responseRule) WithAnyOf(s ...string) interfaces.ResponseBuilder {
	if s == nil || len(s) == 0 {
		return r
	}

	r.with.anyOneOf = s

	return r
}

func (r *responseRule) WithAllOf(s ...string) interfaces.ResponseBuilder {
	if s == nil || len(s) == 0 {
		return r
	}

	r.with.allOf = s

	return r
}

func (r *responseRule) WithExactlyOneOf(s ...string) interfaces.ResponseBuilder {
	if s == nil || len(s) == 0 {
		return r
	}

	r.with.exactlyOneOf = s

	return r
}

func (r *responseRule) WithNoneOf(s ...string) interfaces.ResponseBuilder {
	if s == nil || len(s) == 0 {
		return r
	}

	r.with.noneOf = s

	return r
}

func (r *responseRule) Ignore(s ...string) interfaces.ResponseBuilder {
	if s == nil || len(s) == 0 {
		return r
	}

	r.with.noneOf = append(r.with.noneOf, s...)

	return r
}

func (r *responseRule) Chance(f float64) interfaces.ResponseBuilder {
	r.chance = f

	return r
}

func (r *responseRule) RespondWithOneOf(s ...string) {
	if s == nil || len(s) == 0 {
		return
	}

	r.responses = append(r.responses, s...)
}

func listIdentical(a, b []string) bool {
	if a == nil || b == nil {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for idx := range a {
		if a[idx] != b[idx] {
			return false
		}
	}

	return true
}

func (r *responseRule) Equals(other responseRule) bool {
	if !listIdentical(r.require.exact, other.require.exact) {
		return false
	}

	if !listIdentical(r.require.substring, other.require.substring) {
		return false
	}

	if !listIdentical(r.with.exactlyOneOf, other.with.exactlyOneOf) {
		return false
	}

	if !listIdentical(r.with.anyOneOf, other.with.anyOneOf) {
		return false
	}

	if !listIdentical(r.with.allOf, other.with.allOf) {
		return false
	}

	if !listIdentical(r.with.noneOf, other.with.noneOf) {
		return false
	}

	return true
}
