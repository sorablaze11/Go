package main

import (
    // "fmt"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/go-redis/redis"
    "github.com/gorilla/sessions"
    "golang.org/x/crypto/bcrypt"
    "html/template"
)

var client *redis.Client
var templates *template.Template
var store = sessions.NewCookieStore([]byte("t0p-s3cr3t"))

func main(){
    
    // Redis instance initilization
    client = redis.NewClient(&redis.Options{
        Addr : "localhost:6379",
    })

    templates = template.Must(template.ParseGlob("templates/*.html"))
    r := mux.NewRouter()
    r.HandleFunc("/", AuthRequired(indexGetHandler)).Methods("GET")
    r.HandleFunc("/", AuthRequired(indexPostHandler)).Methods("POST")
    r.HandleFunc("/login", loginGetHandler).Methods("GET")
    r.HandleFunc("/login", loginPostHandler).Methods("POST")
    r.HandleFunc("/register", registerGetHandler).Methods("GET")
    r.HandleFunc("/register", registerPostHandler).Methods("POST")

    // Common file server to handle all static files
    fs := http.FileServer(http.Dir("./static/"))
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
    
    http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

// Middleware for authentication
func AuthRequired(handler http.HandlerFunc) http.HandlerFunc {
    return func (w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session")
        _, ok  := session.Values["username"]
        if !ok {
            http.Redirect(w, r, "/login", 302)
            return
        }
        handler.ServeHTTP(w, r)
    }
}

func indexGetHandler(w http.ResponseWriter, r *http.Request) {
    comments, err := client.LRange("comments", 0, 10).Result()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal server error"))
        return
    }
	templates.ExecuteTemplate(w, "index.html", comments)
}

func indexPostHandler(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    comment := r.PostForm.Get("comment")
    err := client.LPush("comments", comment).Err()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal server error"))
        return
    }
    http.Redirect(w, r, "/", 302)
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
    templates.ExecuteTemplate(w, "login.html", nil)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    username := r.PostForm.Get("username")
    password := r.PostForm.Get("password")
    hash, err := client.Get("user:" + username).Bytes()
    if err == redis.Nil {
        templates.ExecuteTemplate(w, "login.html", "Unknown User")
        return
    } else if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal server error"))
        return
    }
    err = bcrypt.CompareHashAndPassword(hash, []byte(password))
    if err != nil {
        templates.ExecuteTemplate(w, "login.html", "Wrong Password")
        return
    }
    session, _ := store.Get(r, "session")
    session.Values["username"] = username
    session.Save(r, w)
    http.Redirect(w, r, "/", 302)
}

func registerGetHandler(w http.ResponseWriter, r *http.Request){
    templates.ExecuteTemplate(w, "register.html", nil)
}

func registerPostHandler(w http.ResponseWriter, r *http.Request){
    r.ParseForm()
    username := r.PostForm.Get("username")
    password := r.PostForm.Get("password")
    cost := bcrypt.DefaultCost
    hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal server error"))
        return
    }
    err = client.Set("user:" + username, hash, 0).Err()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal server error"))
        return
    }
    http.Redirect(w, r, "/login", 302)
}