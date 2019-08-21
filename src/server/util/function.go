package util

import (
	"reflect"
	"runtime"
	"strings"
)

type Function struct{}

func (f Function) GetFunctionName(i interface{}) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	splitName := strings.Split(fullName, ".")
	if len(splitName) > 1 {
		return splitName[len(splitName)-1]
	}

	return fullName
}
