// Code generated by go generate. DO NOT EDIT.

package opt

import (
	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
)

// ExtractAroundLatLngViaIP returns the first found AroundLatLngViaIPOption from the
// given variadic arguments or nil otherwise.
func ExtractAroundLatLngViaIP(opts ...interface{}) *opt.AroundLatLngViaIPOption {
	for _, o := range opts {
		if v, ok := o.(*opt.AroundLatLngViaIPOption); ok {
			return v
		}
	}
	return nil
}