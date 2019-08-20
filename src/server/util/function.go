package util

import (
	"reflect"
	"runtime"
)

type Function struct{}

func (f Function) GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
