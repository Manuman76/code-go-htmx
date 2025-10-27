package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB
var tmpl *template.Template

type Task struct {
	Id   int
	Task string
	Done bool
}

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:root@(127.0.0.1:3333)/testdb?parseTime=true")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("Database connection established")
}

func main() {
	initDB()
	defer db.Close()

	gRouter := mux.NewRouter()
	gRouter.HandleFunc("/", HomeHandler).Methods("GET")
	gRouter.HandleFunc("/getnewtaskform", getTaskFormHandler).Methods("GET")
	gRouter.HandleFunc("/gettaskupdateform/{id}", getTaskUpdateFormHandler).Methods("GET")
	gRouter.HandleFunc("/tasks/{id}", updateTaskHandler).Methods("PUT", "POST")
	gRouter.HandleFunc("/tasks/{id}", deleteTaskHandler).Methods("DELETE")
	gRouter.HandleFunc("/tasks", fetchTasksHandler).Methods("GET")
	gRouter.HandleFunc("/tasks", addTask).Methods("POST")

	fmt.Println("Server is running on port 3000")
	http.ListenAndServe(":3000", gRouter)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
		http.Error(w, "Error executing template"+err.Error(), http.StatusInternalServerError)
	}
}

func fetchTasksHandler(w http.ResponseWriter, r *http.Request) {
	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todoList", todos)
}

func getTaskFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "addTaskForm", nil)
}

func addTask(w http.ResponseWriter, r *http.Request) {
	task := r.FormValue("task")

	query := "INSERT INTO tasks (task) VALUES (?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		http.Error(w, "Error preparing statement: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(task)
	if err != nil {
		http.Error(w, "Error inserting task into database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	todos, _ := getTasks(db)
	tmpl.ExecuteTemplate(w, "todoList", todos)
}

func getTaskUpdateFormHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId, _ := strconv.Atoi(vars["id"])

	task, err := getTaskByID(db, taskId)
	if err != nil {
		http.Error(w, "Error fetching task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "updateTaskForm", task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId, _ := strconv.Atoi(vars["id"])
	taskItem := r.FormValue("task")
	isDone := r.FormValue("done")

	var taskStatus bool
	switch strings.ToLower(isDone) {
	case "on", "yes":
		taskStatus = true
	case "off", "no":
		taskStatus = false
	default:
		taskStatus = false
	}

	task := Task{
		Id:   taskId,
		Task: taskItem,
		Done: taskStatus,
	}

	query := "UPDATE tasks SET task = ?, done = ? WHERE id = ?"
	result, err := db.Exec(query, task.Task, task.Done, task.Id)
	if err != nil {
		http.Error(w, "Error updating task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "No task found to update with id: "+strconv.Itoa(task.Id), http.StatusNotFound)
		return
	}

	todos, _ := getTasks(db)
	tmpl.ExecuteTemplate(w, "todoList", todos)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId, _ := strconv.Atoi(vars["id"])

	query := "DELETE FROM tasks WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		http.Error(w, "Error preparing delete statement: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(taskId)
	if err != nil {
		http.Error(w, "Error deleting task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "No task found to delete with id: "+strconv.Itoa(taskId), http.StatusNotFound)
		return
	}

	todos, _ := getTasks(db)
	tmpl.ExecuteTemplate(w, "todoList", todos)
}

// Utility Functions
func getTasks(dbPointer *sql.DB) ([]Task, error) {
	query := "SELECT id, task, done FROM tasks"

	rows, err := dbPointer.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.Id, &t.Task, &t.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func getTaskByID(dbPointer *sql.DB, id int) (*Task, error) {
	var t Task
	query := "SELECT id, task, done FROM tasks WHERE id = ?"
	err := dbPointer.QueryRow(query, id).Scan(&t.Id, &t.Task, &t.Done)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no task found for this id %d", id)
		}
		return nil, err
	}
	return &t, nil
}
