package handlers

import (
	"encoding/json"
	"log/slog"

	// "log/slog"
	"net/http"
	"time"

	"github.com/silver-eva/auth_service/auth_service/db"
	"github.com/silver-eva/auth_service/auth_service/lib/jwt"
	"github.com/silver-eva/auth_service/auth_service/lib/utils"
	"github.com/silver-eva/auth_service/auth_service/models"
)

func LoginHandler(db db.PostgresInterface, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginReq models.LoginRequest
		err := json.NewDecoder(r.Body).Decode(&loginReq)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Query the user from the DB
		user, err := db.GetUser(loginReq.Name, loginReq.Password, false)
		if err != nil {
			logger.Error("getting user", slog.Any("error", err))
			resp := models.Response{Status: 401, Message: "wrong creds"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// log in user
		err = db.SetLoggedIn(user.Id, true)
		if err != nil {
			logger.Error("set login", slog.Any("error", err))
			http.Error(w, "Failed to log in", http.StatusInternalServerError)
			return
		}

		// Generate JWT
		token, err := jwt.GenerateJWT(user)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		// Respond with token
		resp := models.Response{RefreshToken: token}
		json.NewEncoder(w).Encode(resp)
	}
}

func AuthHandler(db db.PostgresInterface, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var authReq models.AuthRequest
		err := json.NewDecoder(r.Body).Decode(&authReq)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Decode token
		decodedToken, err := jwt.DecodeJWT(authReq.RefreshToken)
		if err != nil || decodedToken.Expired.Before(time.Now()) {
			resp := models.Response{Status: 401, Message: "token expired"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Verify user in the DB
		user, err := db.GetUser(decodedToken.UserName, decodedToken.UserPass, true)
		if err != nil || !user.IsLoggedIn {
			logger.Error("getting user", slog.Any("error", err))
			resp := models.Response{Status: 409, Message: "something went wrong"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Check roles
		if !utils.Contains(authReq.Roles, user.Role) {
			resp := models.Response{Status: 403, Message: "forbidden"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Generate a new token
		newToken, _ := jwt.GenerateJWT(user)
		resp := models.Response{RefreshToken: newToken}
		json.NewEncoder(w).Encode(resp)
	}
}

func SignupHandler(db db.PostgresInterface, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var signUpReq models.SignUpRequest
		err := json.NewDecoder(r.Body).Decode(&signUpReq)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Create a new user
		user, err := db.CreateUser(signUpReq.Name, signUpReq.Password, signUpReq.Email)
		if err != nil {
			logger.Error("creating user", slog.Any("error", err))
			resp := models.Response{Status: 409, Message: "user already exists"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Generate JWT
		token, _ := jwt.GenerateJWT(user)

		// Respond with success
		resp := models.Response{RefreshToken: token}
		json.NewEncoder(w).Encode(resp)
	}
}

func LogoutHandler(db db.PostgresInterface, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var refreshReq models.RefreshRequest
		err := json.NewDecoder(r.Body).Decode(&refreshReq)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Decode token
		decodedToken, err := jwt.DecodeJWT(refreshReq.RefreshToken)
		if err != nil || decodedToken.Expired.Before(time.Now()) {
			resp := models.Response{Status: 401, Message: "token expired"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Log out user in DB
		err = db.SetLoggedIn(decodedToken.UserId, false)
		if err != nil {
			logger.Error("set logout", slog.Any("error", err))
			http.Error(w, "Failed to log out", http.StatusInternalServerError)
			return
		}

		// Respond with success
		resp := models.Response{Status: 204}
		json.NewEncoder(w).Encode(resp)
	}
}
