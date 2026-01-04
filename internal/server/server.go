package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/konradreiche/pathtrace/internal/analyzer"
	"golang.org/x/exp/trace"
)

type Server struct {
	assets   fs.FS
	analyzer *analyzer.Analyzer
}

func New(assets fs.FS, analyzer *analyzer.Analyzer) *Server {
	return &Server{
		assets:   assets,
		analyzer: analyzer,
	}
}

func (s *Server) Start(addr string) error {
	dist, err := fs.Sub(s.assets, "dist")
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/tasks", http.HandlerFunc(s.handleTasks))
	mux.Handle("/task/{id}", http.HandlerFunc(s.handleTask))
	mux.Handle("/", http.FileServer(http.FS(dist)))
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	ids := make([]uint64, 0, len(s.analyzer.Tasks))
	for id := range s.analyzer.Tasks {
		ids = append(ids, uint64(id))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ids)
}

func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	nodes, ok := s.analyzer.NodesByTask[trace.TaskID(id)]
	if !ok || len(nodes) == 0 {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	root := nodes[0]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flatten(root))
}
