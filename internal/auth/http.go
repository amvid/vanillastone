package auth

import (
	"encoding/json"
	"errors"
	"net/http"
)

// credentials is the JSON body for register and login.
type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleRegister handles POST /register. 201 on success, 409 if taken,
// 400 on bad input.
func (a *Auth) HandleRegister(w http.ResponseWriter, r *http.Request) {
	c, ok := decodeCreds(w, r)
	if !ok {
		return
	}
	switch err := a.Register(c.Username, c.Password); {
	case err == nil:
		w.WriteHeader(http.StatusCreated)
	case errors.Is(err, ErrTaken):
		writeErr(w, http.StatusConflict, "username taken")
	case errors.Is(err, ErrValidation):
		writeErr(w, http.StatusBadRequest, "username 3-20 chars, password 6+ chars")
	default:
		writeErr(w, http.StatusInternalServerError, "server error")
	}
}

// HandleLogin handles POST /login. 200 with {token} on success, 401 otherwise.
func (a *Auth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	c, ok := decodeCreds(w, r)
	if !ok {
		return
	}
	token, err := a.Login(c.Username, c.Password)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "wrong username or password")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func decodeCreds(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST only")
		return credentials{}, false
	}
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return credentials{}, false
	}
	return c, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
