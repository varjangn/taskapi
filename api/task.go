package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/varjangn/taskapi/types"
)

type addTaskBody struct {
	TaskName        string    `json:"task_name"`
	TaskDescription string    `json:"task_description"`
	TaskStatus      string    `json:"task_status"`
	Deadline        time.Time `json:"deadline"`
}

type markDoneBulkBody struct {
	TaskIds []int64 `json:"task_ids"`
}

func (s *APIServer) AddTaskEndpoint(w http.ResponseWriter, r *http.Request) {
	reqUser := r.Context().Value(authUserKey).(*types.User)
	reqBody := new(addTaskBody)
	if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		log.Println("AddTaskError:", err.Error())
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid Request body"})
		return
	}
	if reqBody.TaskName == "" {
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid task name"})
		return
	} else if reqBody.TaskDescription == "" {
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid task description"})
		return
	}
	if !types.ValidTaskStatus.IsValid(reqBody.TaskStatus) {
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid task status, it should be any one of todo,in progress,done"})
		return
	}
	task := types.NewTask(
		reqBody.TaskName, reqBody.TaskDescription, reqBody.TaskStatus, reqBody.Deadline, reqUser,
	)
	newTask, err := s.store.CreateTask(task)
	if err != nil {
		log.Println("AddTaskError:", err.Error())
		WriteJSON(w, http.StatusInternalServerError,
			Map{"error": "failed to create task"})
		return
	}
	WriteJSON(w, http.StatusCreated, newTask)
}

func (s *APIServer) TaskListEndpoint(w http.ResponseWriter, r *http.Request) {
	reqUser := r.Context().Value(authUserKey).(*types.User)
	tasks, err := s.store.TaskList(reqUser, 0, 10)
	if err != nil {
		log.Println("TaskListError:", err.Error())
		WriteJSON(w, http.StatusInternalServerError,
			Map{"error": "failed to list tasks"})
		return
	}
	WriteJSON(w, http.StatusOK, tasks)
}

func (s *APIServer) TaskDetailEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId, err := strconv.ParseInt(vars["taskId"], 10, 64)
	if err != nil {
		log.Println("TaskDetailError:", err.Error())
		WriteJSON(w, http.StatusInternalServerError,
			Map{"error": "failed to fetch task"})
		return
	}

	reqUser := r.Context().Value(authUserKey).(*types.User)
	task, err := s.store.GetTask(taskId, reqUser.Id)
	if err != nil {
		log.Println("TaskDetailError:", err.Error())
		WriteJSON(w, http.StatusInternalServerError,
			Map{"error": "failed to fetch task"})
		return
	}
	WriteJSON(w, http.StatusOK, task)
}

func (s *APIServer) MarkDoneBulk(w http.ResponseWriter, r *http.Request) {
	reqUser := r.Context().Value(authUserKey).(*types.User)
	reqBody := new(markDoneBulkBody)
	if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		log.Println("MarkDoneBulk:", err.Error())
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid Request body"})
		return
	}

	var wg sync.WaitGroup

	for _, taskId := range reqBody.TaskIds {
		wg.Add(1)
		go func(tId, uId int64, w *sync.WaitGroup) {
			defer wg.Done()
			if err := s.store.MarkTaskCompleted(tId, reqUser.Id); err != nil {
				log.Println("MarkDoneBulk:", err.Error())
			}
			log.Println(tId)
		}(taskId, reqUser.Id, &wg)
	}

	wg.Wait()

	WriteJSON(w, http.StatusOK, Map{"success": true})
}

func (s *APIServer) DeleteTaskEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId, err := strconv.ParseInt(vars["taskId"], 10, 64)
	if err != nil {
		log.Println("DeleteTaskEndpoint:", err.Error())
		WriteJSON(w, http.StatusInternalServerError,
			Map{"error": "failed to fetch task"})
		return
	}

	reqUser := r.Context().Value(authUserKey).(*types.User)
	err = s.store.DeleteTask(taskId, reqUser.Id)
	if err != nil {
		log.Println("DeleteTaskEndpoint:", err.Error())
		WriteJSON(w, http.StatusInternalServerError,
			Map{"error": "failed to delete task"})
		return
	}

	WriteJSON(w, http.StatusNoContent, Map{"success": true})
}
