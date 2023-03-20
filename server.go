package gospel

import (
	"io/fs"
	"net/http"
	"strings"
)

type Server struct {
	server     *http.Server
	fileServer http.Handler
	app        *App
}

type PrefixFS struct {
	fs     fs.FS
	prefix string
}

func (f *PrefixFS) Open(name string) (fs.File, error) {
	return f.fs.Open(name[len(f.prefix):])
}

func MakeServer(app *App) *Server {
	return &Server{
		app:        app,
		fileServer: http.FileServer(http.FS(&PrefixFS{fs: app.StaticFiles, prefix: app.StaticPrefix})),
		server: &http.Server{
			Addr: ":8000",
		},
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if strings.HasPrefix(r.URL.Path, s.app.StaticPrefix) {
		s.fileServer.ServeHTTP(w, r)
		return
	}

	ctx := &DefaultContext{}
	elem := s.app.Root(ctx)

	w.Header().Add("content-type", "text/html")

	w.WriteHeader(200)
	w.Write([]byte(elem.RenderElement(ctx)))
}

func (s *Server) Start() error {
	s.server.Handler = s
	go s.server.ListenAndServe()
	return nil
}

func (s *Server) Stop() {

}
