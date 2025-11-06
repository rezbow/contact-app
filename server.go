package contactapp

import (
	"context"
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
	router.Handle("GET /contacts/new", http.HandlerFunc(server.newContactPage))
	router.Handle("POST /contacts/new", http.HandlerFunc(server.newContact))
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticFilesDir)))
	server.Handler = router

	return server
}

func (s *Server) newContactPage(w http.ResponseWriter, r *http.Request) {
	render(w, r.Context(), views.NewContact(models.NewContactForm()))
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
	idStr := r.PathValue("id")
	if id, err := strconv.Atoi(idStr); err == nil {
		contact, err := s.store.GetContact(id)
		if err != nil {
			http.Error(w, "contact not found", http.StatusNotFound)
			return
		}
		render(w, r.Context(), views.ContactDetail(contact))
	} else {
		http.Error(w, "contact not found", http.StatusNotFound)
	}
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
