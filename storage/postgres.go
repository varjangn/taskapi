package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/varjangn/taskapi/types"
)

type PostgresStore struct {
	Db *sql.DB
}

func NewPostgresStorage() (*PostgresStore, error) {
	dbPort, err := strconv.ParseInt(os.Getenv("PGDB_PORT"), 10, 64)
	if err != nil {
		dbPort = 5432
	}
	dbUser := os.Getenv("PGDB_USER")
	dbPass := os.Getenv("PGDB_PASSWORD")
	dbHost := os.Getenv("PGDB_HOST")
	dbName := os.Getenv("PGDB_DBNAME")

	dbURI := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	log.Printf("connecting to %s@%s:%d", dbName, dbHost, dbPort)
	for tryNum := 0; tryNum < 3; tryNum++ {
		db, err := sql.Open("postgres", dbURI)
		if err != nil {
			return nil, err
		}
		err = db.Ping()
		if err == nil {
			log.Println("connected to database")
			return &PostgresStore{
				Db: db,
			}, nil
		}
		log.Println(err.Error())
		time.Sleep(1 * time.Second)
		log.Println("Retrying...")
		continue
	}
	return nil, fmt.Errorf("failed to connect to the database after several tries")
}

func (s *PostgresStore) Init() error {
	if err := s.CreateUsersTable(); err != nil {
		return err
	}
	if err := s.CreateTaskTable(); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) CreateUsersTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		verified BOOLEAN NOT NULL CHECK (verified IN (false, true)),
		first_name TEXT,
		last_name TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := s.Db.Exec(query)
	return err
}

func (s *PostgresStore) CreateTaskTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		task_name TEXT NOT NULL,
		task_description TEXT,
		task_status TEXT,
		deadline TIMESTAMP,
		added_by_user_id INT REFERENCES users(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := s.Db.Exec(query)
	return err
}

func (s *PostgresStore) GetUserId(email string) (int64, error) {
	qry := fmt.Sprintf("SELECT id from users WHERE email='%s'", email)
	row := s.Db.QueryRow(qry)
	var userId int64 = -1

	switch err := row.Scan(&userId); err {
	case sql.ErrNoRows:
		return userId, nil
	case nil:
		return userId, nil
	default:
		return userId, err
	}
}

func (s *PostgresStore) CreateUser(u *types.User) error {
	query := `
		INSERT INTO users(email, password, verified, first_name, last_name)
		VALUES($1, $2, $3, $4, $5)
	`
	_, err := s.Db.Exec(query, u.Email, u.Password, u.Verified, u.FirstName, u.LastName)
	if err != nil {
		return err
	}
	id, err := s.GetUserId(u.Email)
	if err != nil {
		return err
	}
	u.Id = id
	return nil
}

func (s *PostgresStore) GetUser(email string) (*types.User, error) {
	qry := fmt.Sprintf("SELECT * from users WHERE email='%s'", email)
	row := s.Db.QueryRow(qry)
	var user types.User

	switch err := row.Scan(&user.Id, &user.Email, &user.Password, &user.Verified,
		&user.FirstName, &user.LastName, &user.CreatedAt, &user.ModifiedAt); err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &user, nil
	default:
		return nil, err
	}
}

func (s *PostgresStore) CreateTask(t *types.Task) (*types.Task, error) {
	insertQuery := `
		INSERT INTO tasks(task_name, task_description, task_status, deadline, added_by_user_id)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, created_at, modified_at
	`
	row := s.Db.QueryRow(insertQuery, t.TaskName, t.TaskDescription, t.TaskStatus, t.Deadline, t.AddedBy.Id)
	// Scan the returned values into a new Task instance
	var createdTask types.Task
	err := row.Scan(&createdTask.Id, &createdTask.CreatedAt, &createdTask.ModifiedAt)
	if err != nil {
		return nil, err
	}
	// Set the other fields of the createdTask based on the input t
	createdTask.TaskName = t.TaskName
	createdTask.TaskDescription = t.TaskDescription
	createdTask.TaskStatus = t.TaskStatus
	createdTask.Deadline = t.Deadline
	createdTask.AddedBy = t.AddedBy
	return &createdTask, nil
}

func (s *PostgresStore) TaskList(user *types.User, skip, limit int) ([]types.TaskListDetail, error) {
	query := `
		SELECT id, task_name, task_description, task_status, deadline, created_at, modified_at
		FROM tasks
		WHERE added_by_user_id = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3
	`

	rows, err := s.Db.Query(query, user.Id, skip, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []types.TaskListDetail

	for rows.Next() {
		var task types.TaskListDetail
		err := rows.Scan(
			&task.Id,
			&task.TaskName,
			&task.TaskDescription,
			&task.TaskStatus,
			&task.Deadline,
			&task.CreatedAt,
			&task.ModifiedAt,
		)
		if err != nil {
			return nil, err
		}

		// Append the task to the result slice
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *PostgresStore) GetTask(taskId, userId int64) (*types.TaskListDetail, error) {
	query := `
		SELECT id, task_name, task_description, task_status, deadline, created_at, modified_at
		FROM tasks
		WHERE id = $1 and added_by_user_id = $2
	`

	row := s.Db.QueryRow(query, taskId, userId)

	var task types.TaskListDetail
	err := row.Scan(
		&task.Id,
		&task.TaskName,
		&task.TaskDescription,
		&task.TaskStatus,
		&task.Deadline,
		&task.CreatedAt,
		&task.ModifiedAt,
	)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *PostgresStore) MarkTaskCompleted(taskID, userId int64) error {
	query := `
		UPDATE tasks
		SET task_status = 'completed', modified_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND added_by_user_id = $2
	`

	_, err := s.Db.Exec(query, taskID, userId)
	return err
}

func (s *PostgresStore) DeleteTask(taskID, userId int64) error {
	query := `
		DELETE FROM tasks WHERE id = $1 AND added_by_user_id = $2
	`
	_, err := s.Db.Exec(query, taskID, userId)
	return err
}
