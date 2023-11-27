package storage

import (
	"github.com/varjangn/taskapi/types"
)

type Storage interface {
	Init() error
	CreateUsersTable() error
	CreateUser(u *types.User) error
	GetUser(email string) (*types.User, error)
	GetUserId(email string) (int64, error)
	CreateTask(t *types.Task) (*types.Task, error)
	TaskList(user *types.User, skip, limit int) ([]types.TaskListDetail, error)
	GetTask(taskId, userId int64) (*types.TaskListDetail, error)
	MarkTaskCompleted(taskId, userId int64) error
	DeleteTask(taskId, userId int64) error
}
