package requestHandler

import (
	"net/http"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type RequestHandler struct {
	mux *http.ServeMux
}

func NewRequestHandler() *RequestHandler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users/register", RegisterUser)
	return &RequestHandler{mux: mux}
}

func RegisterUser(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	data := req.PostForm
	hash, err := bcrypt.GenerateFromPassword([]byte(data["password"][0]), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Hash to store:", string(hash))
	return
}

func (rH *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rH.mux.ServeHTTP(w, r)
	return
}
