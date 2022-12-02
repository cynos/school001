package tools

import (
	"reflect"
	"strconv"
	"time"
)

func Derefer(val interface{}) interface{} {
	var res interface{}
	switch x := val.(type) {
	case *string:
		res = *x
	case *int:
		res = *x
	case *bool:
		res = *x
	case *time.Time:
		res = *x
	}
	return res
}

func String(val string) *string {
	return &val
}

func Int(val int) *int {
	return &val
}

func Time(val time.Time) *time.Time {
	return &val
}

func StringToInt(val string) int {
	x, _ := strconv.Atoi(val)
	return x
}

func StringToBool(val string) bool {
	x, _ := strconv.ParseBool(val)
	return x
}

func ContainsElementString(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

type Array []interface{}

func (x Array) Includes(e interface{}) bool {
	etype := reflect.TypeOf(e)
	evalue := reflect.ValueOf(e)
	for _, n := range x {
		ntype := reflect.TypeOf(n)
		nvalue := reflect.ValueOf(n)
		if ntype.Kind().String() == etype.Kind().String() {
			if nvalue.String() == evalue.String() {
				return true
			}
		}
	}
	return false
}
