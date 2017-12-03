package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
)

type Prime struct {
	Number  int64 `json:"number"`
	IsPrime bool  `json:"is_prime"`
}

func JudgePrimality(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	i, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		i = 0
	}
	isPrime := big.NewInt(i).ProbablyPrime(1)
	json.NewEncoder(w).Encode(&Prime{i, isPrime})
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/primes/{id}", JudgePrimality).Methods("GET")

	r.Path("/auth/info/googlejwt").Methods("GET").HandlerFunc(authInfoHandler)
	r.Path("/auth/info/googleidtoken").Methods("GET").HandlerFunc(authInfoHandler)
	r.Path("/auth/info/firebase").Methods("GET", "OPTIONS").Handler(corsHandler(authInfoHandler))
	r.Path("/auth/info/auth0").Methods("GET").HandlerFunc(authInfoHandler)

	http.Handle("/", r)
	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		port, _ = strconv.Atoi(portStr)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// corsHandler wraps a HTTP handler and applies the appropriate responses for Cross-Origin Resource Sharing.
type corsHandler http.HandlerFunc

func (h corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		return
	}
	h(w, r)
}

// authInfoHandler reads authentication info provided by the Endpoints proxy.
func authInfoHandler(w http.ResponseWriter, r *http.Request) {
	encodedInfo := r.Header.Get("X-Endpoint-API-UserInfo")
	if encodedInfo == "" {
		w.Write([]byte(`{"id": "anonymous"}`))
		return
	}

	b, err := base64.StdEncoding.DecodeString(encodedInfo)
	if err != nil {
		errorf(w, http.StatusInternalServerError, "Could not decode auth info: %v", err)
		return
	}
	w.Write(b)
}

// errorf writes a swagger-compliant error response.
func errorf(w http.ResponseWriter, code int, format string, a ...interface{}) {
	var out struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	out.Code = code
	out.Message = fmt.Sprintf(format, a...)

	b, err := json.Marshal(out)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Could not format JSON for original message."}`, 500)
		return
	}

	http.Error(w, string(b), code)
}
