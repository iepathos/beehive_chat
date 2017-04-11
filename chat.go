package main // simple chat service

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	// "time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"

	r "gopkg.in/gorethink/gorethink.v3"
)

// database name
var DbName = os.Getenv("DBNAME")

// table name for users service
var TableName = "chat"

// chat models
type Room struct {
	Name  string   `json:"name"`
	Users []string `json:"users"`
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
	Room     string `json:"room"`
	// Timestamp time.Time `json:"timestamp"`
}

// InsertMessage inserts a Message struct into rethinkdb
func InsertMessage(message Message) {
	// connect to db
	session, err := r.Connect(r.ConnectOpts{
		Address: "localhost:28015",
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	db := r.DB(DbName)

	// insert user db
	err = db.Table(TableName).Insert(map[string]interface{}{
		"username": message.Username,
		"message":  message.Message,
		"room":     message.Room,
		// "timestamp": message.Timestamp,
	}).Exec(session)
	if err != nil {
		log.Println(err.Error())
	}
}

// chat views
func CreateMessage(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := req.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var message Message
	err = json.Unmarshal(body, &message)
	if err != nil {
		w.WriteHeader(422)
		log.Println(err.Error())
	}

	InsertMessage(message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(message); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// get messages
func FeedMessages(ws *websocket.Conn) {
	session, err := r.Connect(r.ConnectOpts{
		Address: "localhost:28015",
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	db := r.DB(DbName)

	messages, _ := db.Table(TableName).Changes().Field("new_val").Run(session)
	go func() {
		var msg Message
		for messages.Next(&msg) {
			if msg.Message != "" {
				log.Println("%s: %s", msg.Username, msg.Message)
				if err = websocket.Message.Send(ws, msg); err != nil {
					log.Println("Can't send")
					break
				}
			}
		}
	}()

}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/create", CreateMessage)
	router.Handle("/feed", websocket.Handler(FeedMessages))
	server := &http.Server{
		Addr:    "localhost:3000",
		Handler: router,
	}
	log.Fatal(server.ListenAndServe())
}
