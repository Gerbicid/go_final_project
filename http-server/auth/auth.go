package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
	http_server "go_final_project/http-server"
	"go_final_project/storage/users"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	Authorized bool `json:"authorized"`
	Sum        int  `json:"sum"`
	jwt.StandardClaims
}

func Authorization(w http.ResponseWriter, req *http.Request) {
	var user users.User
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http_server.ResponseJson("Error when reading from req.Body", http.StatusInternalServerError, err, w)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		http_server.ResponseJson("deserialization error", http.StatusInternalServerError, err, w)
		return
	}
	envPassword, err := user.GetPasswordFromEnv()
	passwordVerificationFlag := true
	if err != nil {
		envPassword = ""
		passwordVerificationFlag = false
	}

	if passwordVerificationFlag {
		if user.GetPassword() != envPassword {
			http_server.ResponseJson("wrong password", http.StatusBadRequest, nil, w)
			return
		}
	}
	cookie, err := req.Cookie("token")
	if err == nil {
		tokenStr := cookie.Value
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(envPassword), nil
		})

		if err == nil && token.Valid && claims.ExpiresAt > time.Now().Unix() {
			response := map[string]string{"token": tokenStr}
			answer, err := json.Marshal(response)
			if err != nil {
				http_server.ResponseJson("sirelization error", http.StatusInternalServerError, err, w)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(answer)
			return
		}
	}
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &Claims{
		Authorized: true,
		Sum:        500,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(envPassword))
	if err != nil {
		http_server.ResponseJson("error when creating a new token", http.StatusInternalServerError, err, w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		Expires:  expirationTime,
		HttpOnly: true,
	})
	response := map[string]string{"token": tokenString}
	answer, err := json.Marshal(response)
	if err != nil {
		http_server.ResponseJson("sirelization error", http.StatusInternalServerError, err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(answer)
}

func CheckToken(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var user users.User
		envPassword, err := user.GetPasswordFromEnv()
		if err != nil {
			envPassword = ""
		}
		// парсим JWT-токен из куки и если он действителен пропускаем на следующую страничку
		if len(envPassword) > 0 {
			var cookieToken string
			cookie, err := req.Cookie("token")
			if err != nil {
				http_server.ResponseJson("unauthorized user", http.StatusBadRequest, err, w)
				return
			}
			cookieToken = cookie.Value
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(cookieToken, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(envPassword), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

		}
		next(w, req)
	})
}
