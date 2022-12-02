package report

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

func ToUnderScore(name string) string {
	l := len(name)
	ss := strings.Split(name, "")

	// we just care about the key of idx map,
	// the value of map is meaningless
	idx := make(map[int]int, 1)

	var rs []rune
	for _, s := range name {
		rs = append(rs, []rune(string(s))...)
	}

	for i := l - 1; i >= 0; {
		if unicode.IsUpper(rs[i]) {
			var start, end int
			end = i
			j := i - 1
			for ; j >= 0; j-- {
				if unicode.IsLower(rs[j]) {
					start = j + 1
					break
				}
			}
			// handle the case: "BBC" or "AaBBB" case
			if end == l-1 {
				idx[start] = 1
			} else {
				if start == end {
					// value=1 is meaningless
					idx[start] = 1
				} else {
					idx[start] = 1
					idx[end] = 1
				}
			}
			i = j - 1
		} else {
			i--
		}
	}

	for i := l - 1; i >= 0; i-- {
		ss[i] = strings.ToLower(ss[i])
		if _, ok := idx[i]; ok && i != 0 {
			ss = append(ss[0:i], append([]string{"_"}, ss[i:]...)...)
		}
	}

	return strings.Join(ss, "")
}

func Invoke(any interface{}, name string, args ...interface{}) ([]reflect.Value, error) {
	var (
		method      = reflect.ValueOf(any).MethodByName(name)
		returnvalue = []reflect.Value{}
	)

	if method.Kind() == reflect.Invalid {
		return returnvalue, fmt.Errorf("method %s not found", name)
	}

	methodType := method.Type()
	numIn := methodType.NumIn()

	if numIn > len(args) {
		return returnvalue, fmt.Errorf("method %s must have minimum %d params. Have %d", name, numIn, len(args))
	}

	if numIn != len(args) && !methodType.IsVariadic() {
		return returnvalue, fmt.Errorf("method %s must have %d params. Have %d", name, numIn, len(args))
	}

	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type

		if methodType.IsVariadic() && i >= numIn-1 {
			inType = methodType.In(numIn - 1).Elem()
		} else {
			inType = methodType.In(i)
		}

		argValue := reflect.ValueOf(args[i])
		if !argValue.IsValid() {
			return returnvalue, fmt.Errorf("method %s. param[%d] must be %s. Have %s", name, i, inType, argValue.String())
		}

		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return returnvalue, fmt.Errorf("method %s. param[%d] must be %s. Have %s", name, i, inType, argType)
		}
	}

	return method.Call(in), nil
}

func FindSlice(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func RemoveSlice(s []string, val string) []string {
	i, found := FindSlice(s, val)
	if !found {
		return s
	}
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
