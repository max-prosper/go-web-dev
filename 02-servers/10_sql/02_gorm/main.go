package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
)

// NOT for production! Always handle errors explicitly!
func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}

type Item struct {
	Id          int `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	Title       string
	Description string
	Updated     string `sql:"null"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	
	items := []*Item{}

	db := h.DB.Find(&items)
	err := db.Error
	errPanic(err)	

	err = h.Tmpl.ExecuteTemplate(w, "index.html", struct{
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

	newItem := &Item{
		Title: r.FormValue("title")
		Description: r.FormValue("description"),
	}

	db := h.DB.Create(&newItem)
	err := db.Error
	errPanic(err)

	affected := db.RowsAffected

	fmt.Println("Insert - RowsAffected", affected, "LastInsertId: ", newItem.Id)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	errPanic(err)

	post := &Item{}
	db := h.DB.Find(post, id)
	err = db.Error

	if err == gorm.ErrRecordNotFound {
        fmt.Println("Record not found", id)
	} else {
        errPanic(err)
	}

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

	post := &Item{}
	h.DB.Find(post, id)

	post.Title = r.FormValue("title")
	post.Description = r.FormValue("description")
	post.Updated = "maxprosper"

	db := h.DB.Save(post)
	err = db.Error
	errPanic(err)

	affected := db.RowsAffected

	fmt.Println("Update - RowsAffected", affected)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	errPanic(err)

	db := h.DB.Delete(&Item{Id: id})
	err = db.Error
	errPanic(err)

	affected := db.RowsAffected

	fmt.Println("Delete - RowsAffected", affected)
	w.Header().Set("Content-type", "application/json")
	resp := `"affected": ` + strconv.Itoa(int(affected)) + `}`

	w.Write([]byte(resp))
}

func main() {
	dsn := "root@tcp(localhost:3306)/testdb?"
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := gorm.Open("mysql", dsn)
	db.DB()
	db.Ping()

	handlers := &Handler{
		DB:   db,
		Tmpl: template.Must(template.ParseGlob("../crud_templates/*")),
	}

	// NOTE: authorization and csrf are skipped!!!

	r := mux.NewRouter
	r.HandleFunc("/", handlers.List).Methods("GET")
	r.HandleFunc("/items", handlers.List).Methods("GET")
	r.HandleFunc("/items/new", handlers.Add).Methods("POST")
	r.HandleFunc("/items/{id}", handlers.Edit).Methods("GET")
	r.HandleFunc("/items/{id}", handlers.Update).Methods("POST")
	r.HandleFunc("/items/{id}", handlers.Delete).Methods("DELETE")

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", r)
}