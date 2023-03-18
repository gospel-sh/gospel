package gospel

import (
	"net/http"
)

type Server struct {
	server *http.Server
	app    App
}

func MakeServer(app App) *Server {
	return &Server{
		app: app,
		server: &http.Server{
			Addr: ":8000",
		},
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := Context{}

	elem := s.app(ctx)

	w.Header().Add("content-type", "text/html")

	w.WriteHeader(200)
	w.Write([]byte(elem.Render(ctx)))
}

func (s *Server) Start() error {
	s.server.Handler = s
	go s.server.ListenAndServe()
	return nil
}

func (s *Server) Stop() {

}
