package views

import (
	"net/http"

	"github.com/rezbow/contact-app/models"
)

type FormErrors map[string]string

func (e FormErrors) Set(field, msg string) {
	e[field] = msg
}

func (e FormErrors) Get(field string) string {
	return e[field]
}

type ContactForm struct {
	ID          int
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
	Errors      FormErrors
}

func (c *ContactForm) Valid() bool {
	if c.FirstName == "" {
		c.Errors.Set(ContactFormFirstName, "must not be empty")
	}
	if c.LastName == "" {
		c.Errors.Set(ContactFormLastName, "must not be empty")
	}
	if c.PhoneNumber == "" {
		c.Errors.Set(ContactFormPhone, "must not be empty")
	}
	if c.Email == "" {
		c.Errors.Set(ContactFormEmail, "must not be empty")
	}
	return len(c.Errors) == 0
}

func (c *ContactForm) ToContact() *models.Contact {
	return &models.Contact{
		ID:          c.ID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		Email:       c.Email,
		PhoneNumber: c.PhoneNumber,
	}
}

func ContactFormFromContact(contact *models.Contact) *ContactForm {
	return &ContactForm{
		ID:          contact.ID,
		FirstName:   contact.FirstName,
		LastName:    contact.LastName,
		PhoneNumber: contact.PhoneNumber,
		Email:       contact.Email,
		Errors:      make(FormErrors),
	}
}

func ContactFormFromRequest(r *http.Request) *ContactForm {
	r.ParseForm()
	return &ContactForm{
		FirstName:   r.PostForm.Get(ContactFormFirstName),
		LastName:    r.PostForm.Get(ContactFormLastName),
		Email:       r.PostForm.Get(ContactFormEmail),
		PhoneNumber: r.PostForm.Get(ContactFormPhone),
		Errors:      make(FormErrors),
	}
}
