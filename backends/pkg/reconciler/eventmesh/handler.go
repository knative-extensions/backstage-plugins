package eventmesh

import (
	"context"
	"net/http"
)

func EventMeshHandler(ctx context.Context) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		// TODO: return the mesh here
		w.Write([]byte("Hello World"))
	}
}
