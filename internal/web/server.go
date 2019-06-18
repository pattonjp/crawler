package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pattonjp/crawler/internal"
	"github.com/pattonjp/crawler/internal/indexer"
)

//Server http server for an api
type Server struct {
	router      *chi.Mux
	index       *indexer.Index
	channel     chan *internal.ScrapeRequest
	templates   map[string]*template.Template
	templateDir string
	debug       bool
}

//NewServer creates a new http server
func NewServer(idx *indexer.Index, in chan *internal.ScrapeRequest, dir string, debug bool) (*Server, error) {

	svr := &Server{
		router:      chi.NewRouter(),
		index:       idx,
		channel:     in,
		templateDir: dir,
		debug:       debug,
	}
	svr.routes()
	err := svr.loadTemplates()
	return svr, err

}

func (s *Server) routes() {
	s.router.Get("/", s.home)
	s.router.Get("/index", s.indexForm)
	s.router.Post("/index", s.updateIndex)
	s.router.Route("/api", func(r chi.Router) {
		r.Get("/index", s.search)
		r.Post("/index", s.updateIndex)
		r.Delete("/index", s.resetIndex)
	})
	s.router.Mount("/debug", middleware.Profiler())

}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("q")
	res := s.index.Search(term)
	resp := struct {
		Term    string
		Results []*indexer.WordIdx
	}{
		Term:    term,
		Results: res,
	}

	s.render(w, "home", resp)
}
func (s *Server) indexForm(w http.ResponseWriter, r *http.Request) {
	meta := s.index.Meta()
	s.render(w, "index", meta)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("q")

	res := s.index.Search(term)
	fmt.Printf("Searching for %s\n found %v\n", term, res)
	js, err := json.Marshal(res)
	if err != nil {
		s.handleError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(js); err != nil {
		s.handleError(w)
	}

}

func redirectOrAccept(w http.ResponseWriter, r *http.Request) {
	if r.Referer() != "" {
		http.Redirect(w, r, r.Referer(), http.StatusFound)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) updateIndex(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.handleError(w)
		return
	}

	if r.Form.Get("reset") != "" {
		s.resetIndex(w, r)
		return
	}

	url := r.Form.Get("url")
	fmt.Println("start indexing", url)
	s.channel <- &internal.ScrapeRequest{URL: url}
	redirectOrAccept(w, r)
}
func (s *Server) resetIndex(w http.ResponseWriter, r *http.Request) {
	s.index.Reset()
	redirectOrAccept(w, r)

}

//Close dispose of any server resources
func (s *Server) Close() {

}

//ServeHTTP serves this api on http
func (s *Server) ServeHTTP(addr string) error {
	fmt.Println("Listening on ", addr)
	return http.ListenAndServe(addr, s.router)
}
