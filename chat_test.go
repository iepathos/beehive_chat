// cat_test.go
package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/Sirupsen/logrus"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/iepathos/beehive/rego"
	r "gopkg.in/gorethink/gorethink.v3"
)

func TestCreateMessage(t *testing.T) {
	// lookup user in rethinkdb and make sure it now exists
	session, err := r.Connect(r.ConnectOpts{
		Address: "localhost:28015",
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	rego.CreateDatabase("test")
	rego.CreateTable(TableName)

	url := "/create"
	jsonStr := []byte(`{"username":"Saitama","message":"herro","room":"onepunch"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateMessage)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("CreateMessage handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Check the response body is what we expect.
	reqJSON, err := simplejson.NewFromReader(rr.Body)
	if err != nil {
		t.Errorf("Error while reading request JSON: %s", err)
	}
	username := reqJSON.Get("username").MustString()
	if username != "Saitama" {
		t.Errorf("Expected request JSON response to have username Saitama")
	}
	message := reqJSON.Get("message").MustString()
	if message != "herro" {
		t.Errorf("Expected request JSON response to have message herro but got %s", message)
	}
	room := reqJSON.Get("room").MustString()
	if room != "onepunch" {
		t.Errorf("Expected request JSON response to have room onepunch but got %s", room)
	}

	db := r.DB("test")
	cursor, err := db.Table(TableName).Count().Run(session)
	if err != nil {
		log.Fatalln(err.Error())
	}
	var count int
	cursor.One(&count)
	cursor.Close()
	if count != 1 {
		t.Errorf("Expected RethinkDB chat table to have count of 1")
	}
	rego.DropDatabase("test")
}

// func TestFeedMessages(t *testing.T) {
// 	// lookup user in rethinkdb and make sure it now exists
// 	session, err := r.Connect(r.ConnectOpts{
// 		Address: "localhost:28015",
// 	})
// 	if err != nil {
// 		log.Fatalln(err.Error())
// 	}
// 	rego.CreateDatabase("test")
// 	rego.CreateTable(TableName)

// 	// url := "/feed"
// 	srv := httptest.NewServer(http.HandlerFunc(webs.Handler(FeedMessages)))
// 	u, _ := url.Parse(srv.URL)
// 	u.Scheme = "ws"
// 	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

// 	rego.DropDatabase("test")
// }
