package contactapp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/rezbow/contact-app/models"
	"github.com/rezbow/contact-app/views"
)

const (
	staticFilesDir = http.Dir("./static/")
)

type ContactStore interface {
	GetContacts() []models.Contact
	FilterContacts(string) []models.Contact
	AddContact(models.Contact)
	GetContact(int) (models.Contact, error)
	EditContact(models.Contact) error
	DeleteContact(int) error
}

type Server struct {
	store ContactStore
	http.Handler
}

func NewContactServer(store ContactStore) *Server {
	server := &Server{
		store: store,
	}
	router := http.NewServeMux()
	router.Handle("GET /contacts", http.HandlerFunc(server.getContacts))
	router.Handle("GET /contacts/{id}", http.HandlerFunc(server.getContactDetail))
	router.Handle("GET /contacts/{id}/edit", http.HandlerFunc(server.editContactPage))
	router.Handle("POST /contacts/{id}/edit", http.HandlerFunc(server.editContact))
	router.Handle("POST /contacts/{id}/delete", http.HandlerFunc(server.deleteContact))
	router.Handle("GET /contacts/new", http.HandlerFunc(server.newContactPage))
	router.Handle("POST /contacts/new", http.HandlerFunc(server.newContact))
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticFilesDir)))
	server.Handler = router

	return server
}

func extractId(r *http.Request) (int, error) {
	s := r.PathValue("id")
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("couldn't extract id from path: %w", err)
	}
	return id, nil

}

func (s *Server) deleteContact(w http.ResponseWriter, r *http.Request) {
	id, err := extractId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := s.store.DeleteContact(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	redirect(w, r, "/contacts")
}

func (s *Server) editContactPage(w http.ResponseWriter, r *http.Request) {
	id, err := extractId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	contact, err := s.store.GetContact(id)
	if err != nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}
	render(w, r.Context(), views.ContactEdit(models.NewContactForm(&contact)))
}

func (s *Server) editContact(w http.ResponseWriter, r *http.Request) {
	id, err := extractId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	form := models.ContactFormFromRequest(r)
	form.Contact.ID = id
	if !form.Validate() {
		render(w, r.Context(), views.ContactEdit(form))
		return
	}
	err = s.store.EditContact(form.Contact)
	if err != nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}
	redirect(w, r, fmt.Sprintf("/contacts/%d", id))
}

func (s *Server) newContactPage(w http.ResponseWriter, r *http.Request) {
	render(w, r.Context(), views.NewContact(models.NewContactForm(nil)))
}

func (s *Server) newContact(w http.ResponseWriter, r *http.Request) {
	form := models.ContactFormFromRequest(r)
	if !form.Validate() {
		render(w, r.Context(), views.NewContact(form))
		return
	}
	s.store.AddContact(form.GetContact())
	redirect(w, r, "/contacts")
}

func (s *Server) getContactDetail(w http.ResponseWriter, r *http.Request) {
	id, err := extractId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	contact, err := s.store.GetContact(id)
	if err != nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}
	render(w, r.Context(), views.ContactDetail(contact))
}

func (s *Server) getContacts(w http.ResponseWriter, r *http.Request) {
	var contacts []models.Contact
	q := r.URL.Query().Get("q")
	if q == "" {
		contacts = s.store.GetContacts()
	} else {
		contacts = s.store.FilterContacts(q)
	}
	data := views.ContactsViewModel{
		Contacts: contacts,
		Query:    q,
	}
	render(w, r.Context(), views.Contacts(data))
}

func render(w http.ResponseWriter, ctx context.Context, content templ.Component) {
	if err := views.Base(content, "title").Render(ctx, w); err != nil {
		log.Println(err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}
