package main

import (
	"fmt"
	"github/websocket/pubsub"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ps = pubsub.PubSub{}
)

func autoId() string {
	return uuid.Must(uuid.NewV4(), nil).String()
}

func websockethandler(w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	//catch the new client
	fmt.Println("server: new client is connected.")

	client := &pubsub.Client{
		Id:         autoId(),
		Connection: conn,
	}
	//add client to array
	ps.AddClient(client)

	for {
		//read message from client
		messageType, p, err := conn.ReadMessage()
		if err != nil {

			ps.RemoveClient(client)
			log.Printf("server: error on reading message: %v", err)

			return
		}

		//msg := []byte("Hi Client i am server.")
		//fmt.Printf("server: new message from client: %s", p)
		//send message to client
		//if err := conn.WriteMessage(messageType, msg); err != nil {
		//	log.Println("server: error on writting message: %v", err)
		//	return
		//}

		//receive from client
		ps.HandleReceiveMessage(client, messageType, p)

	}
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//payload := map[string]interface{}{
		//	"message": "Hello GO",
		//}
		//w.Header().Set("Content-Type", "application/json")
		//w.Header().Set("Content-Security-Policy", "default-src *; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' 'unsafe-eval' http://www.google.com")
		//json.NewEncoder(w).Encode(payload)

		http.ServeFile(w, r, "static")
	})

	http.HandleFunc("/ws", websockethandler)

	log.Fatal(http.ListenAndServe(":3000", nil))
	log.Printf("Server is running at https://localhost:3000")
}
