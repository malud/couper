// Code generated by go generate. DO NOT EDIT.

package opt

import "encoding/json"

// KeepDiacriticsOnCharactersOption is a wrapper for an KeepDiacriticsOnCharacters option parameter. It holds
// the actual value of the option that can be accessed by calling Get.
type KeepDiacriticsOnCharactersOption struct {
	value string
}

// KeepDiacriticsOnCharacters wraps the given value into a KeepDiacriticsOnCharactersOption.
func KeepDiacriticsOnCharacters(v string) *KeepDiacriticsOnCharactersOption {
	return &KeepDiacriticsOnCharactersOption{v}
}

// Get retrieves the actual value of the option parameter.
func (o *KeepDiacriticsOnCharactersOption) Get() string {
	if o == nil {
		return ""
	}
	return o.value
}

// MarshalJSON implements the json.Marshaler interface for
// KeepDiacriticsOnCharactersOption.
func (o KeepDiacriticsOnCharactersOption) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface for
// KeepDiacriticsOnCharactersOption.
func (o *KeepDiacriticsOnCharactersOption) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.value = ""
		return nil
	}
	return json.Unmarshal(data, &o.value)
}

// Equal returns true if the given option is equal to the instance one. In case
// the given option is nil, we checked the instance one is set to the default
// value of the option.
func (o *KeepDiacriticsOnCharactersOption) Equal(o2 *KeepDiacriticsOnCharactersOption) bool {
	if o == nil {
		return o2 == nil || o2.value == ""
	}
	if o2 == nil {
		return o == nil || o.value == ""
	}
	return o.value == o2.value
}

// KeepDiacriticsOnCharactersEqual returns true if the two options are equal.
// In case of one option being nil, the value of the other must be nil as well
// or be set to the default value of this option.
func KeepDiacriticsOnCharactersEqual(o1, o2 *KeepDiacriticsOnCharactersOption) bool {
	return o1.Equal(o2)
}