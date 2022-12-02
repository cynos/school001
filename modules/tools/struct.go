package tools

import (
	"reflect"

	"gorm.io/gorm"
)

func ForEachStruct(data interface{}, callback func(key string, val interface{})) {
	var (
		paramsType  = reflect.TypeOf(data)
		paramsValue = reflect.ValueOf(data)
	)
	for i := 0; i < paramsType.NumField(); i++ {
		field := paramsType.Field(i)
		value := paramsValue.Field(i)
		switch value.Interface().(type) {
		case gorm.Model:
			for ii := 0; ii < value.NumField(); ii++ {
				x := field.Type.Field(ii)
				y := value.Field(ii)
				callback(x.Name, y.Interface())
			}
		default:
			callback(field.Name, value.Interface())
		}
	}
}

func GetDataStuct(data interface{}, tag string) map[string]interface{} {
	var (
		dataType  = reflect.TypeOf(data)
		dataValue = reflect.ValueOf(data)
		result    = make(map[string]interface{})
	)

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		value := dataValue.Field(i)
		if tagname, ok := field.Tag.Lookup(tag); ok {
			if tagname != "" {
				result[tagname] = value.Interface()
			} else {
				result[field.Name] = value.Interface()
			}
		} else {
			result[field.Name] = value.Interface()
		}
	}

	return result
}
