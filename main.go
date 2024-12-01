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

const (
	jwtSignKey = "your_jwt_secret_key"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health-check", healthCheckHandler)
	mux.HandleFunc("/users/register", userRegisterHandler)
	mux.HandleFunc("/users/login", userLoginHandler)
	mux.HandleFunc("/users/profile", userProfileHandler)

	log.Printf("Server is listening on :8080...")
	server := http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(server.ListenAndServe())
}

func healthCheckHandler(writer http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(writer, `{"message": "everything is OK!"}`)
}

func userRegisterHandler(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(writer, "Only POST method is allowed")
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	var uReq userservice.RegisterRequest
	err = json.Unmarshal(data, &uReq)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	mysqlRepo := mysql.New()
	userSvc := userservice.New(mysqlRepo, jwtSignKey)

	_, err = userSvc.Register(uReq)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(`{"message": "User registered successfully"}`))
}

func userLoginHandler(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(writer, "Only POST method is allowed")
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	var lReq userservice.LoginRequest
	err = json.Unmarshal(data, &lReq)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	mysqlRepo := mysql.New()
	userSvc := userservice.New(mysqlRepo, jwtSignKey)

	resp, err := userSvc.Login(lReq)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	data, err = json.Marshal(resp)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(data)
}

func userProfileHandler(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(writer, "Only GET method is allowed")
		return
	}

	pReg := userservice.ProfileRequest{UserID: 0}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	err = json.Unmarshal(data, &pReg)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	mysqlRepo := mysql.New()
	userSvc := userservice.New(mysqlRepo, jwtSignKey)
	resp, err := userSvc.Profile(pReg)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	data, err = json.Marshal(resp)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(data)
}
