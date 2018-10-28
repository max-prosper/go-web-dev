package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// NOT for production! Always handle errors explicitly!
func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}

type Handler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items := []*Item{}

	rows, err := h.DB.Query("SELECT id, title, update FROM items")
	errPanic(err)

	for rows.Next() {
		post := &Item{}
		err = rows.Scan(&post.Id, &post.Title, &post.Updated)
		errPanic(err)
		items = append(items, post)
	}

	rows.Close()

	err = h.Tmpl.ExecuteTemplate(w, "index.html", struct {
		Items []*Item
	}{
		Items: items,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {

	// NOTE: validation is skipped!

	result, err := h.DB.Exec(
		"INSERT INTO items (`title`, `description`) VALUES (?, ?)",
		r.FormValue("title"),
		r.FormValue("description"),
	)
	errPanic(err)

	affected, err := result.RowsAffected()
	errPanic(err)

	lastID, err := result.LastInsertId()
	errPanic(err)

	fmt.Println("Insert - RowsAffected", affected, "LastInsertId: ", lastID)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	errPanic(err)

	post := &Item{}
	row := h.DB.QueryRow("SELECT id, title, updated, description FROM items WHERE id = ?", id)
	err = row.Scan(&post.Id, &post.Title, &post.Updated, &post.Description)
	errPanic(err)

	err = h.Tmpl.ExecuteTemplate(w, "edit.html", post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	errPanic(err)

	// NOTE: validation is skipped!
	result, err := h.DB.Exec("UPDATE items SET"+
		"`title` = ?"+
		",`description` = ?"+
		",`updated` = ?"+
		"WHERE id = ?",
		r.FormValue("title"),
		r.FormValue("description"),
		"maxprosper",
		id,
	)
	errPanic(err)

	affected, err := result.RowsAffected()
	errPanic(err)

	fmt.Println("Update - RowsAffected", affected)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	errPanic(err)

	result, err := h.DB.Exec(
		"DELETE FROM items WHERE id = ?",
		id,
	)
	errPanic(err)

	affected, err := result.RowsAffected()
	errPanic(err)

	fmt.Println("Delete - RowsAffected", affected)

	w.Header().Set("Content-type", "application/json")
	resp := `{"affected": ` + strconv.Itoa(int(affected)) + `}`
	w.Write([]byte(resp))
}

func main() {
	dsn := "root@tcp(localhost:3306)/testdb?"
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)
	db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	handlers := &Handler{
		DB:   db,
		Tmpl: template.Must(template.ParseGlob("../crud_templates/*")),
	}

	// NOTE: authorization and csrf are skipped!!!

	r := mux.NewRouter
	r.HandleFunc("/", handlers.List).Methods("GET")
	r.HandleFunc("/items", handlers.List).Methods("GET")
	r.HandleFunc("/items/new", handlers.AddForm).Methods("GET")
	r.HandleFunc("/items/new", handlers.Add).Methods("POST")
	r.HandleFunc("/items/{id}", handlers.Edit).Methods("GET")
	r.HandleFunc("/items/{id}", handlers.Update).Methods("POST")
	r.HandleFunc("/items/{id}", handlers.Delete).Methods("DELETE")

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", r)
}
