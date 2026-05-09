package application

import "reflect"

// DeleteRegisteredBindingMethodIDForTesting removes a previously registered
// stable binding ID from the package-level map. Only for use in tests.
func DeleteRegisteredBindingMethodIDForTesting(method any) {
	registeredBindingMethodIDs.Delete(reflect.ValueOf(method).Pointer())
}
