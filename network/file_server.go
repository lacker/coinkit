package network

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"coinkit/data"
)

type FileServer struct {
	store *data.DataStore
}

func NewFileServer(store *data.DataStore) *FileServer {
	return &FileServer{
		store: store,
	}
}

func (fs *FileServer) ServeForever(port int) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimLeft(r.URL.Path, "/")
		log.Printf("handling [%s]", key)
		value, ok := fs.store.Get(key)
		if !ok {
			fmt.Fprintf(w, "no data")
		} else {
			fmt.Fprintf(w, value)
		}
	}
	http.HandleFunc("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%d"), nil)
}
