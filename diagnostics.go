package diagnostics

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Diagnostic struct {
	message  string
	data     map[string]interface{}
	innerErr *error
	inner    *Diagnostic
	isErr    bool
}

func None() Diagnostic {
	return Diagnostic{}
}

func (d Diagnostic) Error() string {
	if d.inner != nil {
		return fmt.Sprintf("%s: %s", d.message, d.inner.Error())
	}
	if d.innerErr != nil {
		return (*d.innerErr).Error()
	}
	return d.message
}

func (d Diagnostic) Is(err error) bool {
	if d.innerErr != nil {
		if errors.Is(*d.innerErr, err) {
			return true
		}
	}
	if d.inner != nil {
		return d.inner.Is(err)
	}

	return false
}
func (d Diagnostic) As(val interface{}) bool {
	if d.innerErr != nil {
		if errors.As(*d.innerErr, val) {
			return true
		}
	}
	if d.inner != nil {
		return d.inner.As(val)
	}

	return false
}

func (d Diagnostic) UserFacing(depth ...int) string {
	start := 1
	if len(depth) > 0 {
		start = depth[0]
	}

	var dataRep string
	if d.data != nil {
		dataRep = fmt.Sprintf(": %s", printable(d.data))
	}

	if d.innerErr != nil {
		return fmt.Sprintf("%s%s", *d.innerErr, dataRep)
	}
	if d.inner != nil {
		indent := ""
		for i := 0; i < start; i++ {
			indent = fmt.Sprintf("    %s", indent)
		}
		return fmt.Sprintf("%s%s\n%s┗━ %s", d.message, dataRep, indent, d.inner.UserFacing(start+1))
	}

	return fmt.Sprintf("%s%s", d.message, dataRep)
}

func printable(data map[string]interface{}) string {
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var s []string
	for _, key := range keys {
		value := data[key]
		if val, ok := value.(string); ok {
			s = append(s, fmt.Sprintf("%s = %q", key, val))
		} else {
			s = append(s, fmt.Sprintf("%s = %v", key, value))
		}
	}

	return strings.Join(s, ", ")
}

func (d Diagnostic) Wrap(message string) Diagnostic {
	return Diagnostic{
		isErr:    true,
		message:  message,
		data:     nil,
		innerErr: nil,
		inner:    &d,
	}
}

func (d Diagnostic) WithData(key string, value interface{}) Diagnostic {
	newData := d.data
	if newData == nil {
		newData = map[string]interface{}{}
	}
	newData[key] = value
	return Diagnostic{
		isErr:    true,
		message:  d.message,
		data:     newData,
		innerErr: d.innerErr,
		inner:    d.inner,
	}
}

func (d Diagnostic) WithAllData(m map[string]interface{}) Diagnostic {
	newData := d.data
	if newData == nil {
		newData = m
	} else {
		for k, v := range m {
			newData[k] = v
		}
	}
	return Diagnostic{
		isErr:    true,
		message:  d.message,
		data:     newData,
		innerErr: d.innerErr,
		inner:    d.inner,
	}
}

func (d Diagnostic) IsErr() bool {
	return d.isErr
}

func (d Diagnostic) Err() error {
	if d.isErr {
		return d
	}
	return nil
}

func FromString(message string) Diagnostic {
	return Diagnostic{
		message: message,
		isErr:   true,
	}
}

func FromErr(err error) Diagnostic {
	if err == nil {
		return None()
	}
	return Diagnostic{
		innerErr: &err,
		isErr:    true,
	}
}
