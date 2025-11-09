package contactapp

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/rezbow/contact-app/models"
)

type InMemoryStore struct {
	contacts []models.Contact
	idSeq    int
}

func (s *InMemoryStore) get(by func(c models.Contact) bool) *models.Contact {
	for _, contact := range s.contacts {
		if by(contact) {
			return &contact
		}
	}
	return nil
}
func (s *InMemoryStore) nextId() int {
	s.idSeq++
	return s.idSeq
}

func totalPage(total int) int {
	return int(math.Ceil(float64(total) / float64(10)))
}

func paged(s []models.Contact, page int) []models.Contact {
	var contacts []models.Contact
	start := (page - 1) * 10
	for idx := start; idx < len(s) && len(contacts) != 10; idx++ {
		contacts = append(contacts, s[idx])
	}
	return contacts
}

func (s *InMemoryStore) GetContacts(page int) ([]models.Contact, int) {
	return paged(s.contacts, page), totalPage(len(s.contacts))
}

func (s *InMemoryStore) FilterContacts(q string, page int) ([]models.Contact, int) {
	var contacts []models.Contact
	for _, contact := range s.contacts {
		if strings.Contains(contact.FirstName, q) || strings.Contains(contact.LastName, q) {
			contacts = append(contacts, contact)
		}
	}
	return paged(contacts, page), totalPage(len(contacts))
}

func (s *InMemoryStore) AddContact(contact models.Contact) error {
	if s.DuplicateEmail(contact.Email, 0) {
		return ErrDuplicateEmail
	}
	contact.ID = s.nextId()
	s.contacts = append(s.contacts, contact)
	return nil
}

func (s *InMemoryStore) GetContact(id int) (models.Contact, error) {
	for _, contact := range s.contacts {
		if contact.ID == id {
			return contact, nil
		}
	}
	return models.Contact{}, errors.New("contact not found ")
}

func (s *InMemoryStore) EditContact(contact models.Contact) error {
	if s.DuplicateEmail(contact.Email, contact.ID) {
		return ErrDuplicateEmail
	}
	for idx, c := range s.contacts {
		if c.ID == contact.ID {
			s.contacts[idx] = contact
			return nil
		}
	}
	return ErrNotFound
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
		return ErrNotFound
	}
	s.contacts = contacts
	return nil

}

func (s *InMemoryStore) DuplicateEmail(email string, id int) bool {
	contactWithSameEmail := s.get(func(c models.Contact) bool {
		return c.Email == email
	})
	if contactWithSameEmail != nil && contactWithSameEmail.ID != id {
		return true
	}
	return false
}

// expensive call WOWO
func (s *InMemoryStore) Count() int {
	time.Sleep(time.Second * 5)
	return len(s.contacts)
}

func NewinMemoryStore() *InMemoryStore {

	return &InMemoryStore{
		contacts: []models.Contact{
			{ID: 1, FirstName: "Jack", LastName: "Jackson", Email: "jack@jaskcons.com", PhoneNumber: "213214"},
			{ID: 2, FirstName: "John", LastName: "Doe", Email: "john@doe.com", PhoneNumber: "123142"},
			{ID: 3, FirstName: "Arthur", LastName: "Morgan", Email: "artur@morgan.com", PhoneNumber: "213214"},
		},
		idSeq: 1,
	}

}
