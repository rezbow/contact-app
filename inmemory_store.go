package contactapp

import (
	"errors"
	"strings"

	"github.com/rezbow/contact-app/models"
)

type InMemoryStore struct {
	contacts []models.Contact
	idSeq    int
}

func (s *InMemoryStore) nextId() int {
	s.idSeq++
	return s.idSeq
}

func (s *InMemoryStore) GetContacts() []models.Contact {
	return s.contacts
}

func (s *InMemoryStore) FilterContacts(q string) []models.Contact {
	var contacts []models.Contact
	for _, contact := range s.GetContacts() {
		if strings.Contains(contact.FirstName, q) || strings.Contains(contact.LastName, q) {
			contacts = append(contacts, contact)
		}
	}
	return contacts
}

func (s *InMemoryStore) AddContact(contact models.Contact) {
	contact.ID = s.nextId()
	s.contacts = append(s.contacts, contact)
}

func (s *InMemoryStore) GetContact(id int) (models.Contact, error) {
	for _, contact := range s.contacts {
		if contact.ID == id {
			return contact, nil
		}
	}
	return models.Contact{}, errors.New("contact not found ")
}

func (s *InMemoryStore) EditContact(c models.Contact) error {
	for idx, contact := range s.contacts {
		if contact.ID == c.ID {
			s.contacts[idx] = c
			return nil
		}
	}
	return errors.New("contact not found ")

}

func (s *InMemoryStore) DeleteContact(id int) error {
	var contacts []models.Contact
	found := false
	for _, contact := range s.contacts {
		if contact.ID == id {
			found = true
			continue
		}
		contacts = append(contacts, contact)
	}
	if !found {
		return errors.New("contact not found ")
	}
	s.contacts = contacts
	return nil

}

func NewinMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		contacts: make([]models.Contact, 0),
	}
}
