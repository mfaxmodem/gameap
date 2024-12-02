package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mfaxmodem/gameap/repository/mysql"
	"github.com/mfaxmodem/gameap/service/userservice"
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

	// بارگیری کلید خصوصی
	privateKeyPath := "keys/private-key.pem"
	privateKey, err := userservice.LoadPrivateKey(privateKeyPath)
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
	}
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
	userSvc := userservice.New(mysqlRepo, privateKey)

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
	// بارگیری کلید خصوصی
	privateKeyPath := "keys/private-key.pem"
	privateKey, err := userservice.LoadPrivateKey(privateKeyPath)
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
	}
	if req.Method != http.MethodPost {
		http.Error(writer, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var lReq userservice.LoginRequest
	if err := json.NewDecoder(req.Body).Decode(&lReq); err != nil {
		http.Error(writer, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	mysqlRepo := mysql.New()
	userSvc := userservice.New(mysqlRepo, privateKey)

	resp, err := userSvc.Login(lReq)
	if err != nil {
		http.Error(writer, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnauthorized)
		return
	}

	if err := json.NewEncoder(writer).Encode(resp); err != nil {
		http.Error(writer, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
	}
}

func userProfileHandler(writer http.ResponseWriter, req *http.Request) {
	// بارگیری کلید خصوصی
	privateKeyPath := "keys/private-key.pem"
	privateKey, err := userservice.LoadPrivateKey(privateKeyPath)
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
	}
	if req.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(writer, "Only GET method is allowed")
		return
	}

	auth := req.Header.Get("Authorization")
	claims, err := ParsJWT(auth)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte(fmt.Sprintf(`{"error":"Invalid token: %v"}`, err)))
		return
	}

	mysqlRepo := mysql.New()

	userSvc := userservice.New(mysqlRepo, privateKey)
	resp, err := userSvc.Profile(userservice.ProfileRequest{UserID: claims.UserID})
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(data)
}

func ParsJWT(tokenStr string) (*userservice.Claims, error) {
	tokenStr = strings.TrimSpace(strings.Replace(tokenStr, "Bearer ", "", 1))

	publicKeyPath := "keys/public-key.pem"
	publicKey, err := userservice.LoadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %v", err)
	}

	token, err := jwt.ParseWithClaims(tokenStr, &userservice.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(*userservice.Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
