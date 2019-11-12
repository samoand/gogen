package mock

import "runtime"

func RuntimeCallerSupport0() (string, bool) {
	_, filename, _, ok := runtime.Caller(0)
	return filename, ok
}

func RuntimeCallerSupport1() (string, bool) {
	_, filename, _, ok := runtime.Caller(1)
	return filename, ok
}
