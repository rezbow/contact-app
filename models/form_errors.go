package models

type FormErrors map[string]string

func (e FormErrors) Set(field, msg string) {
	e[field] = msg
}

func (e FormErrors) Get(field string) string {
	return e[field]
}
