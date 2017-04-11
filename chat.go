// simple chat service
import time

// chat models
type Room struct {
	name string
	users []string
}

type Message struct {
	username string
	message string
	timestamp time.Time
	room string
}



// chat views

// create message
// get messages
