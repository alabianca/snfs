package util

import (
	"encoding/json"
	"log"
	"net/http"
)

func Message(status int, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")

	if status, ok := data["status"]; ok {
		code := status.(int)
		w.WriteHeader(code)
	}
	log.Println("res", data)
	json.NewEncoder(w).Encode(data)
}
