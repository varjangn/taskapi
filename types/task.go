package types

import (
	"strings"
	"time"
)

type ValidTaskStatusType [3]string

var ValidTaskStatus ValidTaskStatusType = [3]string{"todo", "in progress", "done"}

func (v *ValidTaskStatusType) IsValid(status string) bool {
	for _, s := range v {
		if s == strings.ToLower(status) {
			return true
		}
	}
	return false
}

type Task struct {
	Id              int64     `json:"id"`
	TaskName        string    `json:"task_name"`
	TaskDescription string    `json:"task_description"`
	TaskStatus      string    `json:"task_status"`
	Deadline        time.Time `json:"deadline"`
	AddedBy         *User     `json:"user"`
	CreatedAt       time.Time `json:"created_at"`
	ModifiedAt      time.Time `json:"modified_at"`
}

type TaskListDetail struct {
	Id              int64     `json:"id"`
	TaskName        string    `json:"task_name"`
	TaskDescription string    `json:"task_description"`
	TaskStatus      string    `json:"task_status"`
	Deadline        time.Time `json:"deadline"`
	CreatedAt       time.Time `json:"created_at"`
	ModifiedAt      time.Time `json:"modified_at"`
}

func NewTask(name, desc, status string, deadline time.Time, addedBy *User) *Task {
	return &Task{
		Id:              0,
		TaskName:        name,
		TaskDescription: desc,
		TaskStatus:      status,
		Deadline:        deadline,
		AddedBy:         addedBy,
		CreatedAt:       time.Now().In(time.UTC),
		ModifiedAt:      time.Now().In(time.UTC),
	}
}

func (t *Task) IsDone() bool {
	return strings.ToLower(t.TaskStatus) == "completed"
}
