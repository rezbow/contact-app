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
	"github.com/sebdah/goldie"
)

type StubContactStore struct {
	contacts []models.Contact
	addCalls []models.Contact
	idSeq    int
}

func (s *StubContactStore) nextId() int {
	s.idSeq++
	return s.idSeq
}

func (s *StubContactStore) AddContact(contact models.Contact) {
	contact.ID = s.nextId()
	s.addCalls = append(s.addCalls, contact)
}

func (s *StubContactStore) GetContacts() []models.Contact {
	return s.contacts
}

func (s *StubContactStore) GetContact(id int) (models.Contact, error) {
	for _, contact := range s.contacts {
		if contact.ID == id {
			return contact, nil
		}
	}
	return models.Contact{}, errors.New("contact not found ")
}

func (s *StubContactStore) FilterContacts(q string) []models.Contact {
	var contacts []models.Contact
	for _, contact := range s.GetContacts() {
		if strings.Contains(contact.FirstName, q) || strings.Contains(contact.LastName, q) {
			contacts = append(contacts, contact)
		}
	}
	return contacts
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
	f := url.Values{}
	f.Set(models.ContactFormFirstName, contact.FirstName)
	f.Set(models.ContactFormLastName, contact.LastName)
	f.Set(models.ContactFormPhone, contact.PhoneNumber)
	f.Set(models.ContactFormEmail, contact.Email)
	req, _ := http.NewRequest(http.MethodPost, "/contacts/new", strings.NewReader(f.Encode()))
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
}
