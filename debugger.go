package main

import "log"
import "net/http"
import "github.com/gorilla/websocket"

var vm *VM

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// use conn to send and receive
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err = conn.WriteMessage(messageType, p); err != nil {
			return
		}
	}
}

func Debug(thevm *VM) {
	vm = thevm
	http.HandleFunc("/debug", handler)
	http.Handle("/", http.FileServer(http.Dir("www")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
