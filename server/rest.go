package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

/*
	Routes and associated methods:
	/			GET								// Used only to serve a static placeholder
	/notifs		GET, POST, PUT, DELETE			// Used to get, create, and update pending notifs
	/devices	GET, POST, DELETE				// Used to get the devices within a watchgroup
	/enroll		GET								// Used to get keys for enrolling new devices to a watchgroup
*/

// main function to start the server and init REST routes, returns relevant errors if things go wrong
func StartREST(listeningPort int, e chan any, db *sql.DB) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", recoverHandler(e, serveRoot))


	// NOTIFS
	mux.HandleFunc("GET /notifs/pending", recoverHandler(e, func(w http.ResponseWriter, req *http.Request) {
		getPendingNotifs(w, req, db)
	}))
	mux.HandleFunc("GET /notifs/complete", recoverHandler(e, func(w http.ResponseWriter, req *http.Request) {
		getCompleteNotifs(w, req, db)
	}))
	mux.HandleFunc("POST /notifs", recoverHandler(e, func(w http.ResponseWriter, req *http.Request) {
		createNotifs(w, req, db)
	}))

	mux.HandleFunc("PUT /notifs", recoverHandler(e, func(w http.ResponseWriter, req *http.Request) {
		updateNotifs(w, req, db)
	}))
	mux.HandleFunc("DELETE /notifs", recoverHandler(e, func(w http.ResponseWriter, req *http.Request) {
		deleteNotifs(w, req, db)
	}))
	


	//run server (makes sure it blocks so server keeps running)
	addr := ":" + strconv.Itoa(listeningPort)
	if err := http.ListenAndServe(addr, mux); err != nil {e <- err}
}

// wraps handler to catch panics and push them to the error channel - it wont work the prev way you had it 
func recoverHandler(e chan any, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("PANIC in webserver, pushing upstream")
			e <- r
		}
	}()
		next(w, req)
	}
}


//serve the static placeholder
func serveRoot(w http.ResponseWriter, req *http.Request) {
	//i think http.ServeFile would be better here but it won't have as explicit logging but idk if that's a big deal
		// if you do wanna use it just uncomment the line below and comment out line 53-65
	//http.ServeFile(w, req, "static/index.html")
	body, err := os.ReadFile("static/index.html")
	if err != nil {
		fmt.Println("Couldn't find static file to serve on root:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
	fmt.Println("Served static root to", req.RemoteAddr)
}

func getPendingNotifs(w http.ResponseWriter, req *http.Request, db *sql.DB) {
	notifs, err := GetPendingNotifs(db)
	if err != nil {
		fmt.Println("Failed to fetch pending notifs : ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": notifs})
	fmt.Println("GET /notifs/pending served to", req.RemoteAddr)
}

func getCompleteNotifs(w http.ResponseWriter, req *http.Request, db *sql.DB) {
	notifs, err := GetCompletedNotifs(db)
	if err != nil {
		fmt.Println("Failed to fetch complete notifs : ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": notifs})
	fmt.Println("GET /notifs/complete served to", req.RemoteAddr)
}

func createNotifs(w http.ResponseWriter, req *http.Request, db *sql.DB) {
	var body struct {
		Header      string `json:"header"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		fmt.Println("(createNotifs) invalid JSON:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	id, err := CreateNotif(db, body.Header, body.Description)
	if err != nil {
		fmt.Println("Failed to create notif : ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"data": Notification{ID: int(id), Header: body.Header, Description: body.Description},
	})
	fmt.Println("Created notification with id", id)
}

func updateNotifs(w http.ResponseWriter, req *http.Request, db *sql.DB) {
	id, ok := queryInt(w, req, "id")
	if !ok {return}
	var body struct {
		Header      string `json:"header"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		fmt.Println("(updateNotifs) invalid JSON:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	//todo for lou - add validation for status
	if err := UpdateNotif(db, id, body.Header, body.Description, body.Status); err != nil {
		fmt.Println("Failed to update notif : ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": "updated"})
	fmt.Println("Updated notification with id", id)
}

func deleteNotifs(w http.ResponseWriter, req *http.Request, db *sql.DB) {
	id, ok := queryInt(w, req, "id")
	if !ok {return}
	if err := DeleteNotif(db, id); err != nil {
		fmt.Println("failed to delete notif : ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	fmt.Println("Deleted notification with id", id)
}

// reads a query param like ?id=5, parses it as an int then returns (value, true) if param is valid 
// if param is missing/invalid it writes a 400
// used in updateNotifs + deleteNotifs to get the notif id from the url
func queryInt(w http.ResponseWriter, req *http.Request, key string) (int, bool) {
	paramStr := req.URL.Query().Get(key)
	if paramStr == "" {
		http.Error(w, key+" required", http.StatusBadRequest)
		return 0, false
	}
	n, err := strconv.Atoi(paramStr)
	if err != nil {
		http.Error(w, "invalid "+key, http.StatusBadRequest)
		return 0, false
	}
	return n, true
}

//helper to write json
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
