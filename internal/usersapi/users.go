package usersapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vatsal3003/viswals/internal/consts"
	"github.com/vatsal3003/viswals/internal/database"
	"github.com/vatsal3003/viswals/internal/service/userservice"
	"github.com/vatsal3003/viswals/models"
	"go.uber.org/zap"
)

type API struct {
	DB     *database.Database
	Logger *zap.Logger
}

func New(db *database.Database, logger *zap.Logger) *API {
	return &API{
		DB:     db,
		Logger: logger,
	}
}

func (api *API) InitRoutes() {
	http.HandleFunc("GET /users", api.GetAllUsers)
	http.HandleFunc("GET /users/{userID}", api.GetUser)
	http.HandleFunc("GET /users/sse", api.GetAllUsersSSE)
}

func (api *API) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Query().Get("first_name")
	lname := r.URL.Query().Get("last_name")

	filters := make(map[string]string)

	if fname != "" {
		filters["first_name"] = fname
	} else if lname != "" {
		filters["last_name"] = lname
	}

	users, err := userservice.GetAllUsers(api.DB, filters)
	if err != nil {
		api.Logger.Error("failed to get all users from database:" + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(models.Response{
		Status: consts.StatusSuccess,
		Data:   users,
	})
	if err != nil {
		api.Logger.Error("failed to encode users to JSON: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

func (api *API) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")

	user, err := userservice.GetUser(api.DB, userID)
	if err != nil {
		api.Logger.Error("failed to get user from database:" + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(models.Response{
		Status: consts.StatusSuccess,
		Data:   user,
	})
	if err != nil {
		api.Logger.Error("failed to encode user to JSON: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (api *API) GetAllUsersSSE(w http.ResponseWriter, r *http.Request) {
	limit := 25

	qLimit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if qLimit < limit {
		limit = qLimit
	}

	// Set essential headers
	w.Header().Set("Content-Type", "text/event-stream") // its mandatory for SSE
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// flusher to send data immediately to client using Flush function
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming Unsupported", http.StatusInternalServerError)
		return
	}

	users, err := userservice.GetAllUsers(api.DB, nil)
	if err != nil {
		api.Logger.Error("failed to get all users from database:" + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if limit > len(users) {
		limit = len(users)
	}

	for i := 0; i < limit; i++ {
		data, err := json.Marshal(users[i])
		if err != nil {
			api.Logger.Error("failed to marshal user:" + err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "data: "+string(data)+"\n\n")
		flusher.Flush()
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Fprint(w, "data: END\n\n")
	flusher.Flush()
}
