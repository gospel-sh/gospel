package gospel

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

type Server struct {
	fs         fs.FS
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

func computeETag(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:]))
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

	fs := &PrefixFS{
		fs:     &MultiFS{append([]fs.FS{JS}, app.StaticFiles...)},
		prefix: app.StaticPrefix,
	}

	return &Server{
		app:        app,
		fs:         fs,
		fileServer: http.FileServer(http.FS(fs)),
		server: &http.Server{
			Addr: ":8001",
		},
	}
}

var makeInMemoryStore = MakeInMemoryStoreRegistry()
var makeCookieStore = MakeCookieStoreRegistry()

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if strings.HasPrefix(r.URL.Path, s.app.StaticPrefix) {

		filePath := filepath.Join(".", r.URL.Path)

		// Check if the file exists and is not a directory
		file, err := s.fs.Open(filePath)

		if err == nil {
			defer file.Close()
			fileInfo, err := file.Stat()
			if err == nil && !fileInfo.IsDir() {
				// Read file contents to compute the ETag
				fileContents, err := ioutil.ReadAll(file)
				if err == nil {
					etag := computeETag(fileContents)
					w.Header().Set("ETag", etag)
					w.Header().Set("Cache-Control", "max-age=3600, stale-while-revalidate=3600")
				}
			}
		}

		ifNoneMatch := r.Header.Get("If-None-Match")
		eTag := r.Header.Get("ETag")

		// If the ETag in the request matches the computed ETag, return 304 Not Modified
		if ifNoneMatch != "" && eTag != "" && ifNoneMatch == eTag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		s.fileServer.ServeHTTP(w, r)
		return
	}

	// we make a persistent store for the session
	persistentStore := makeCookieStore(r)
	store := MakeStore(persistentStore)
	ctx := MakeDefaultContext(r, w, store)

	// we set up the router (it adds itself to the context)...
	router := MakeRouter(ctx)

	elem := ctx.Execute(s.app.Root)

	store.Finalize()
	persistentStore.Finalize(w)

	if redirectedTo := router.RedirectedTo(); redirectedTo != "" && (redirectedTo != r.URL.Path || r.Method != http.MethodGet) {
		http.Redirect(w, r, redirectedTo, 302)
		return
	}

	if ctx.RespondWith() != nil {
		ctx.RespondWith()(ctx, w)
		return
	}

	w.Header().Add("content-type", "text/html")
	w.WriteHeader(ctx.StatusCode())
	w.Write([]byte(elem.RenderElement()))

}

func (s *Server) Start() error {
	s.server.Handler = s
	go s.server.ListenAndServe()
	return nil
}

func (s *Server) Stop() {
	// to do: implement stop
}
