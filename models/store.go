package models

type ContactStore interface {
	GetByEmail(string) (Contact, error)
	Add(Contact) error
	Edit(Contact) error
	Delete(int) error
	Get(int) (Contact, error)
	All() []Contact
}
