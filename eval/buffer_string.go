// Code generated by "stringer -type=BufferOption -output=./buffer_string.go"; DO NOT EDIT.

package eval

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BufferNone-0]
	_ = x[BufferRequest-1]
	_ = x[BufferResponse-2]
	_ = x[JSONParseRequest-4]
	_ = x[JSONParseResponse-8]
}

const (
	_BufferOption_name_0 = "BufferNoneBufferRequestBufferResponse"
	_BufferOption_name_1 = "JSONParseRequest"
	_BufferOption_name_2 = "JSONParseResponse"
)

var (
	_BufferOption_index_0 = [...]uint8{0, 10, 23, 37}
)

func (i BufferOption) String() string {
	switch {
	case i <= 2:
		return _BufferOption_name_0[_BufferOption_index_0[i]:_BufferOption_index_0[i+1]]
	case i == 4:
		return _BufferOption_name_1
	case i == 8:
		return _BufferOption_name_2
	default:
		return "BufferOption(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
