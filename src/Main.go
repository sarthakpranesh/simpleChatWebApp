package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// defining a message body
type Message struct {
        Email           string  `json: "email"`
        Username        string  `json: "username"`
        Message         string  `json: "message"`
}


var (
	// connected
	clients = make(map[*websocket.Conn]bool)
	
	// broadcast
	broadcast = make(chan Message)
	
	// Configure an upgrader - it takes in a http connection and upgrades it too socket connection
	upgrader = websocket.Upgrader{}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}


		// closing connection when function returns
		defer ws.Close()

		// Register our new client
		clients[ws] = true

		for {
			var msg Message
		
			// read the message as json and map it to the Message Object
			err := ws.ReadJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				delete(clients, ws)
				break
			}
			broadcast <- msg
		}
}

func handleMessages() {
	for {
		// Grab the next message from braodcast channel
		msg := 	<-broadcast
		log.Printf("In the Message distributor")
		for client := range clients {
			log.Printf("Sending to client")
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	// creating a simple file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// route for configuring socket connecion
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	// starting the server and making sure there are no errors
	log.Println("http server started");
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal("Listen and Serve: ", err)
	}
}

