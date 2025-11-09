package contactapp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/rezbow/contact-app/models"
	"github.com/rezbow/contact-app/views"
)

var (
	ErrDuplicateEmail = errors.New("email is taken")
	ErrNotFound       = errors.New("contact not found")
)

const (
	staticFilesDir = http.Dir("./static/")
)

type ContactStore interface {
	GetContacts(page int) ([]models.Contact, int)
	FilterContacts(string, int) ([]models.Contact, int)
	AddContact(models.Contact) error
	GetContact(int) (models.Contact, error)
	EditContact(models.Contact) error
	DeleteContact(int) error
	DuplicateEmail(email string, contactId int) bool
	Count() int
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
	router.Handle("DELETE /contacts", http.HandlerFunc(server.deleteBulkContact))
	router.Handle("GET /contacts/{id}", http.HandlerFunc(server.getContactDetail))
	router.Handle("GET /contacts/{id}/edit", http.HandlerFunc(server.editContactPage))
	router.Handle("POST /contacts/{id}/edit", http.HandlerFunc(server.editContact))
	router.Handle("DELETE /contacts/{id}", http.HandlerFunc(server.deleteContact))
	router.Handle("GET /contacts/new", http.HandlerFunc(server.newContactPage))
	router.Handle("POST /contacts/new", http.HandlerFunc(server.newContact))
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticFilesDir)))

	router.Handle("GET /contacts/{id}/email", http.HandlerFunc(server.checkEmail))
	router.Handle("GET /contacts/count", http.HandlerFunc(server.getCount))

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

func (s *Server) getCount(w http.ResponseWriter, r *http.Request) {
	renderString(w, fmt.Sprintf("total count is %d", s.store.Count()))
}

// checks if a given email is valid for a contact
func (s *Server) checkEmail(w http.ResponseWriter, r *http.Request) {
	id, err := extractId(r)
	if err != nil {
		return
	}
	email := r.URL.Query().Get("email")
	if email == "" {
		return
	}
	if s.store.DuplicateEmail(email, id) {
		w.Write([]byte(ErrDuplicateEmail.Error()))
	}
}

// /contacts
func (s *Server) deleteBulkContact(w http.ResponseWriter, r *http.Request) {
	idsStr := r.URL.Query()["selected_id"]
	log.Println(idsStr)
	for _, str := range idsStr {
		id, err := strconv.Atoi(str)
		if err != nil || id <= 0 {
			continue
		}
		if s.store.DeleteContact(id) != nil {
			continue
		}
	}
	contacts, totalPages := s.store.GetContacts(1)
	viewModel := views.ContactsViewModel{
		Contacts:   contacts,
		Pagination: views.NewPagination(1, totalPages, r.URL),
	}
	render(w, r.Context(), views.Contacts(viewModel))
}

func (s *Server) deleteContact(w http.ResponseWriter, r *http.Request) {
	id, err := extractId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := s.store.DeleteContact(id); err != nil {
		switch err {
		case ErrNotFound:
			http.NotFound(w, r)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	if isInlineDelete(r) {
		log.Println("client is using inline delete")
		renderString(w, "")
		return
	}
	redirect(w, r, "/contacts")
}

func isInlineDelete(r *http.Request) bool {
	return r.Header.Get("HX-Trigger") == "delete-link"
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
	render(w, r.Context(), views.ContactEdit(views.ContactFormFromContact(&contact)))
}

func (s *Server) editContact(w http.ResponseWriter, r *http.Request) {
	id, err := extractId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	form := views.ContactFormFromRequest(r)
	form.ID = id
	if !form.Valid() {
		render(w, r.Context(), views.ContactEdit(form))
		return
	}
	if err := s.store.EditContact(*form.ToContact()); err != nil {
		switch err {
		case ErrDuplicateEmail:
			form.Errors.Set(views.ContactFormEmail, err.Error())
			render(w, r.Context(), views.ContactEdit(form))
		case ErrNotFound:
			http.NotFound(w, r)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	redirect(w, r, fmt.Sprintf("/contacts/%d", id))
}

func (s *Server) newContactPage(w http.ResponseWriter, r *http.Request) {
	render(w, r.Context(), views.NewContact(&views.ContactForm{}))
}

func (s *Server) newContact(w http.ResponseWriter, r *http.Request) {
	form := views.ContactFormFromRequest(r)
	if !form.Valid() {
		render(w, r.Context(), views.NewContact(form))
		return
	}
	if err := s.store.AddContact(*form.ToContact()); err != nil {
		switch err {
		case ErrDuplicateEmail:
			form.Errors.Set(views.ContactFormEmail, err.Error())
			render(w, r.Context(), views.NewContact(form))
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
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
	var (
		contacts  []models.Contact
		totalPage int
	)
	page, _ := extractPaginationData(r.URL.Query())
	q := r.URL.Query().Get("q")
	if q == "" {
		contacts, totalPage = s.store.GetContacts(page)
	} else {
		contacts, totalPage = s.store.FilterContacts(q, page)
	}
	data := views.ContactsViewModel{
		Contacts:   contacts,
		Query:      q,
		Pagination: views.NewPagination(page, totalPage, r.URL),
	}
	if isActiveSearch(r) {
		log.Println("client hit us with a active search request")
		renderPartial(w, r.Context(), views.Rows(contacts, data.Pagination))
		return
	}
	render(w, r.Context(), views.Contacts(data))
}

func render(w http.ResponseWriter, ctx context.Context, content templ.Component) {
	if err := views.Base(content, "title").Render(ctx, w); err != nil {
		log.Println(err)
	}
}

func renderPartial(w http.ResponseWriter, ctx context.Context, content templ.Component) {
	if err := content.Render(ctx, w); err != nil {
		log.Println(err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func isActiveSearch(r *http.Request) bool {
	return r.Header.Get("HX-Trigger") == "search"
}

func renderString(w http.ResponseWriter, data any) {
	fmt.Fprintf(w, "%v", data)
}
