package handler

type Funcs struct {
	IsActiveNavLink func(string, string) string
}

func IsActiveNavLink() func(string, string) string {
	return func(currentLink, path string) string {
		if currentLink == path {
			return "active"
		} else {
			return ""
		}
	}
}
