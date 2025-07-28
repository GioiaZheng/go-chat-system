package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// RegisterRequest matches the OpenAPI schema
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
	Gender   string `json:"gender,omitempty"`
}

// RegisterResponse matches the OpenAPI schema
type RegisterResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    []RegisterUserData `json:"data"`
}

type RegisterUserData struct {
	User  models.User `json:"user"`
	Token string      `json:"token,omitempty"` // 实际就是 user.ID
}

func (rt *_router) doRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse request
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	// Create user model
	newUser := models.User{
		Username: req.Username,
		Email:    req.Email,
		Name:     req.Name,
		Gender:   req.Gender,
	}

	// Call database
	createdUser, err := rt.db.CreateUser(newUser, req.Password)
	if err != nil {
		if isDuplicateError(err) {
			rt.writeErrorResponse(w, http.StatusConflict, "User already exists")
		} else {
			rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to register user")
		}
		return
	}

	// Token 就是 user.ID
	token := createdUser.ID

	// Prepare response
	response := RegisterResponse{
		Code:    http.StatusCreated,
		Message: "User registered successfully",
		Data: []RegisterUserData{
			{
				User:  createdUser,
				Token: token,
			},
		},
	}

	rt.writeJSONResponse(w, http.StatusCreated, response)
}

// Helper functions
func (rt *_router) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    statusCode,
		"message": message,
	})
}

func (rt *_router) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func isDuplicateError(err error) bool {
	// You can implement proper SQLite error detection here
	return false
}
