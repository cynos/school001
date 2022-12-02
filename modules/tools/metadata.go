package tools

import "net/url"

func SetMetaData(data map[string]string) string {
	d := url.Values{}
	for k, v := range data {
		d.Set(k, v)
	}
	return d.Encode()
}

func GetMetaData(data string, key string) string {
	d, err := url.ParseQuery(data)
	if err != nil {
		return ""
	}
	return d.Get(key)
}
