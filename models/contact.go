package models

import "net/http"

const (
	ContactFormFirstName = "first_name"
	ContactFormLastName  = "last_name"
	ContactFormEmail     = "email"
	ContactFormPhone     = "phone"
)

type FormErrors map[string]string

func (e FormErrors) Set(field, msg string) {
	e[field] = msg
}

func (e FormErrors) Get(field string) string {
	return e[field]
}

type Contact struct {
	ID          int
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
}

type ContactForm struct {
	Errors FormErrors
	Contact
}

func (f *ContactForm) Validate() bool {
	if f.FirstName == "" {
		f.Errors.Set(ContactFormFirstName, "must not be empty")
	}
	if f.LastName == "" {
		f.Errors.Set(ContactFormLastName, "must not be empty")
	}
	if f.Email == "" {
		f.Errors.Set(ContactFormEmail, "must not be empty")
	}
	if f.PhoneNumber == "" {
		f.Errors.Set(ContactFormPhone, "must not be empty")
	}
	return len(f.Errors) == 0
}

func (f *ContactForm) GetContact() Contact {
	return f.Contact
}

func NewContactForm() *ContactForm {
	return &ContactForm{
		Errors: make(FormErrors),
	}
}

func ContactFormFromRequest(req *http.Request) *ContactForm {
	req.ParseForm()
	contact := Contact{
		FirstName:   req.FormValue(ContactFormFirstName),
		LastName:    req.FormValue(ContactFormLastName),
		PhoneNumber: req.FormValue(ContactFormPhone),
		Email:       req.FormValue(ContactFormEmail),
	}
	return &ContactForm{
		Contact: contact,
		Errors:  make(FormErrors),
	}
}
