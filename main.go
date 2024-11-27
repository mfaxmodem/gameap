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
	mux.HandleFunc("/user/register", userRegisterHandler)
	log.Printf("Server is listening on :8080...")
	server := http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(server.ListenAndServe())
}

func userRegisterHandler(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		fmt.Fprintf(writer, "Only POST method is allowed")
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		writer.Write([]byte(
			fmt.Sprintf(`{"error": "#{error.Error()}"}`),
		))
	}
	var uReq userservice.RegisterRequest
	err = json.Unmarshal(data, &uReq)
	if err != nil {
		writer.Write([]byte(
			fmt.Sprintf(`{"error": "#{error.Error()}"}`),
		))
		return
	}
	mysqlRepo := mysql.New()

	userSvc := userservice.New(mysqlRepo)

	_, err = userSvc.Register(uReq)
	if err != nil {
		writer.Write([]byte(
			fmt.Sprintf(`{"error": "#{error.Error()}"}`),
		))
		return
	}
	writer.Write([]byte(`{"message": "User registered successfully"}`))
}
