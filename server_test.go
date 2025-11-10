package contactapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rezbow/contact-app/models"
	"github.com/rezbow/contact-app/views"
	"github.com/sebdah/goldie"
)

type StubContactStore struct {
	contacts    []models.Contact
	addCalls    []models.Contact
	editCalls   []models.Contact
	deleteCalls []int
	idSeq       int
}

func (s *StubContactStore) nextId() int {
	s.idSeq++
	return s.idSeq
}
func (s *StubContactStore) Count() int {
	return 0
}

func (s *StubContactStore) AddContact(contact models.Contact) error {
	contact.ID = s.nextId()
	s.addCalls = append(s.addCalls, contact)
	return nil
}

func (s *StubContactStore) GetContacts(page int) ([]models.Contact, int) {
	return s.contacts, 0
}

func (s *StubContactStore) GetContact(id int) (models.Contact, error) {
	for _, contact := range s.contacts {
		if contact.ID == id {
			return contact, nil
		}
	}
	return models.Contact{}, errors.New("contact not found ")
}

func (s *StubContactStore) FilterContacts(q string, page int) ([]models.Contact, int) {
	return s.contacts, 0
}

func (s *StubContactStore) DuplicateEmail(email string, id int) bool {
	return true
}

func (s *StubContactStore) EditContact(contact models.Contact) error {
	s.editCalls = append(s.editCalls, contact)
	return nil
}

func (s *StubContactStore) DeleteContact(id int) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func newGetRequest(path string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	return req
}

func newGetRequestWithQuery(path string, q string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	req.URL.RawQuery = fmt.Sprintf("q=%s", q)
	return req
}

func newContactRequest(contact models.Contact) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/contacts/new", strings.NewReader(contactToForm(contact)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func assertGolden(t *testing.T, data []byte) {
	t.Helper()
	goldie.Assert(t, t.Name(), data)
}

func assertCode(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected code %d, got %d", want, got)
	}
}

func assertRedirect(t testing.TB, res *httptest.ResponseRecorder, url string) {
	t.Helper()
	assertCode(t, res.Code, http.StatusSeeOther)
	if res.Header().Get("Location") != url {
		t.Errorf("expected Location header to be %q, got %q", url, res.Header().Get("Location"))
	}
}

func TestServer(t *testing.T) {
	store := &StubContactStore{
		contacts: []models.Contact{
			{ID: 1, FirstName: "Chris", LastName: "Jackson", PhoneNumber: "92213", Email: "ChrisJackson@email.com"},
			{ID: 2, FirstName: "John", LastName: "Doe", PhoneNumber: "754639", Email: "JohnDoe@email.com"},
		},
		idSeq: 2,
	}
	server := NewContactServer(store)
	t.Run("request to contacts returns all contacts", func(t *testing.T) {
		req := newGetRequest("/contacts")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)
		assertCode(t, res.Code, http.StatusOK)
		assertGolden(t, res.Body.Bytes())
	})

	t.Run("request to contacts with search query 'Chris' returns correct contacts", func(t *testing.T) {
		req := newGetRequestWithQuery("/contacts", "Chris")
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)

		assertCode(t, res.Code, http.StatusOK)
		assertGolden(t, res.Body.Bytes())
	})

	t.Run("request to an undefined route returns 404 code", func(t *testing.T) {
		req := newGetRequest("/undefined")
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		assertCode(t, res.Code, http.StatusNotFound)
	})

	t.Run("get new contact page(form) ", func(t *testing.T) {
		req := newGetRequest("/contacts/new")
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		assertCode(t, res.Code, http.StatusOK)
		assertGolden(t, res.Body.Bytes())
	})

	t.Run("add new contact", func(t *testing.T) {
		contact := models.Contact{
			ID:          store.idSeq + 1,
			FirstName:   "Reza",
			LastName:    "Bolhasani",
			PhoneNumber: "093223323",
			Email:       "rez@gmail.com",
		}
		req := newContactRequest(contact)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)
		assertRedirect(t, res, "/contacts")

		if len(store.addCalls) != 1 {
			t.Fatalf("got %d call to AddContact, wanted %d ", len(store.addCalls), 1)
		}

		if store.addCalls[0] != contact {
			t.Errorf("%v added to store, wanted %v", store.addCalls[0], contact)
		}
	})

	t.Run("error on invalid new contact form ", func(t *testing.T) {
		contact := models.Contact{
			FirstName:   "",
			LastName:    "Bolhasani",
			PhoneNumber: "",
			Email:       "rez@gmail.com",
		}
		req := newContactRequest(contact)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)
		assertGolden(t, res.Body.Bytes())
	})

	t.Run("get contact detail", func(t *testing.T) {
		id := 1
		req := newGetRequest(fmt.Sprintf("/contacts/%d", id))
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)

		assertCode(t, res.Code, http.StatusOK)
		assertGolden(t, res.Body.Bytes())
	})

	t.Run("return 404 for missing contact", func(t *testing.T) {
		req := newGetRequest(fmt.Sprintf("/contacts/%d", 32329))
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)

		assertCode(t, res.Code, http.StatusNotFound)
		assertGolden(t, res.Body.Bytes())
	})

	t.Run("edit contact page", func(t *testing.T) {
		req := newGetRequest(fmt.Sprintf("/contacts/%d/edit", 1))
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)

		assertCode(t, res.Code, http.StatusOK)
		assertGolden(t, res.Body.Bytes())

	})

	t.Run("edit contact", func(t *testing.T) {
		contact := store.contacts[0]
		contact.FirstName = "Charles"
		contact.LastName = "White"
		contact.PhoneNumber = "091020"
		contact.Email = "Charles@white.com"

		req := editContactRequest(contact)
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)

		assertRedirect(t, res, fmt.Sprintf("/contacts/%d", contact.ID))

		if len(store.editCalls) != 1 {
			t.Fatalf("call to edit must be %d, got %d", 1, len(store.editCalls))
		}

		if store.editCalls[0] != contact {
			t.Errorf("edit contact recived wrong argument, got %v, wanted %v", store.editCalls[0], contact)
		}

	})

	t.Run("delete contact", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/contacts/%d", 1), nil)
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)

		assertRedirect(t, res, "/contacts")

		if len(store.deleteCalls) != 1 {
			t.Fatalf("got %d call to delete, wanted %d", len(store.editCalls), 1)
		}

		if store.deleteCalls[0] != 1 {
			t.Errorf("got %d as argument to delete, wanted %d", store.deleteCalls[0], 1)
		}
	})
}

func editContactRequest(contact models.Contact) *http.Request {
	body := strings.NewReader(contactToForm(contact))
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/contacts/%d/edit", contact.ID), body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func contactToForm(contact models.Contact) string {
	f := url.Values{}
	f.Set(views.ContactFormFirstName, contact.FirstName)
	f.Set(views.ContactFormLastName, contact.LastName)
	f.Set(views.ContactFormPhone, contact.PhoneNumber)
	f.Set(views.ContactFormEmail, contact.Email)
	return f.Encode()
}
