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

type MultiFS struct {
	fileSystems []fs.FS
}

func (m *MultiFS) Open(name string) (fs.File, error) {
	var err error
	var file fs.File
	for _, fileSystem := range m.fileSystems {
		if file, err = fileSystem.Open(name); err == nil {
			return file, nil
		}
	}
	return nil, err
}

func MakeServer(app *App) *Server {
	return &Server{
		app:        app,
		fileServer: http.FileServer(http.FS(&PrefixFS{fs: &MultiFS{append([]fs.FS{JS}, app.StaticFiles...)}, prefix: app.StaticPrefix})),
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

	ctx := MakeDefaultContext()

	// we set up the router...
	router := MakeRouter(r)
	router.SetContext(ctx)

	elem := ctx.Execute(s.app.Root)

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
