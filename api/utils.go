package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Map map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func ExtractJWTToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("invalid header")
	}
	authParts := strings.Split(authHeader, " ")
	if len(authParts) != 2 || strings.ToLower(authParts[0]) != "bearer" {
		return "", fmt.Errorf("invalid header")
	}
	return authParts[1], nil
}
