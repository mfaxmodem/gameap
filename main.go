package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mfaxmodem/gameap/repository/mysql"
	"github.com/mfaxmodem/gameap/service/userservice"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health-heck", healthCheckHandler)
	mux.HandleFunc("/users/register", userRegisterHandler)
	mux.HandleFunc("/users/login", userLoginHandler)

	log.Printf("Server is listening on :8080...")
	server := http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(server.ListenAndServe())
}
func healthCheckHandler(writer http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(writer, `{"message": "everything is OK!"}`)
}

func userRegisterHandler(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		fmt.Fprintf(writer, "Only POST method is allowed")
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		writer.Write([]byte(
			fmt.Sprintf(`{"error":"%s"}`, err.Error()),
		))
	}

	var uReq userservice.RegisterRequest
	err = json.Unmarshal(data, &uReq)
	if err != nil {
		writer.Write([]byte(
			fmt.Sprintf(`{"error":"%s"}`, err.Error()),
		))
		return
	}

	mysqlRepo := mysql.New()
	userSvc := userservice.New(mysqlRepo)

	_, err = userSvc.Register(uReq)
	if err != nil {
		writer.Write([]byte(
			fmt.Sprintf(`{"error":"%s"}`, err.Error()),
		))
		return
	}
	writer.Write([]byte(`{"message": "User registered successfully"}`))
}

func userLoginHandler(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		fmt.Fprintf(writer, "Only POST method is allowed")
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		writer.Write([]byte(
			fmt.Sprintf(`{"error":"%s"}`, err.Error()),
		))
	}

	var lReq userservice.LoginRequest
	err = json.Unmarshal(data, &lReq)
	if err != nil {
		writer.Write([]byte(
			fmt.Sprintf(`{"error":"%s"}`, err.Error()),
		))
		return
	}

	mysqlRepo := mysql.New()
	userSvc := userservice.New(mysqlRepo)

	_, err = userSvc.Login(lReq)
	if err != nil {
		writer.Write([]byte(
			fmt.Sprintf(`{"error":"%s"}`, err.Error()),
		))
		return
	}

	// TODO: Implement login logic
	writer.Write([]byte(`{"message": "User logged in successfully"}`))
}
