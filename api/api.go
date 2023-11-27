package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/varjangn/taskapi/storage"
)

type APIServer struct {
	listenAddr string
	store      storage.Storage
}

func NewAPIServer(store storage.Storage) *APIServer {
	listenAddr := fmt.Sprintf("%s:%s", os.Getenv("API_HOST"), os.Getenv("API_PORT"))
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "API is running",
		"v":      "v1",
	})
}

func (s *APIServer) Run() error {
	if err := s.store.Init(); err != nil {
		log.Fatal(err)
	}

	prefix := "/api/v1/"

	r := mux.NewRouter()
	r.Use(LoggingMiddleware)
	r.HandleFunc(prefix, Method(index, "GET"))
	r.HandleFunc(prefix+"user/register/", Method(s.RegisterEndpoint, "POST"))
	r.HandleFunc(prefix+"user/login/", Method(s.LoginEndpoint, "POST"))
	r.HandleFunc(prefix+"user/profile/", Method(JWTAuth(s.ProfileEndpoint, s.store), "GET"))
	r.HandleFunc(prefix+"task/add/", Method(JWTAuth(s.AddTaskEndpoint, s.store), "POST"))
	r.HandleFunc(prefix+"task/list/", Method(JWTAuth(s.TaskListEndpoint, s.store), "GET"))
	r.HandleFunc(prefix+"task/detail/{taskId}", Method(JWTAuth(s.TaskDetailEndpoint, s.store), "GET"))
	r.HandleFunc(prefix+"task/{taskId}/remove/", Method(JWTAuth(s.DeleteTaskEndpoint, s.store), "DELETE"))
	r.HandleFunc(prefix+"task/mark-done-bulk/", Method(JWTAuth(s.MarkDoneBulk, s.store), "POST"))

	log.Println("server running on", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, r)
}
