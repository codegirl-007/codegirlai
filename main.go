package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

// Connected clients and usernames
var clients = make(map[*websocket.Conn]string) // Maps WebSocket connection to username
var usernames = make(map[string]bool)          // Tracks active usernames
var broadcast = make(chan Message)
var mutex = &sync.Mutex{} // To handle concurrent map access

// Message struct
type Message struct {
	Type     string   `json:"type"`     // "message", "typing", "user_list", or "admin_status"
	Content  string   `json:"content"`  // Message content
	Sender   string   `json:"sender"`   // Sender's username
	Receiver string   `json:"receiver"` // Receiver's username
	Users    []string `json:"users"`    // List of connected users (for "user_list" messages)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer ws.Close()

	// Get username from query parameters
	username := r.URL.Query().Get("username")
	if username == "" {
		ws.WriteMessage(websocket.TextMessage, []byte("Error: Username is required."))
		return
	}

	mutex.Lock()
	// Check if username is unique
	if usernames[username] {
		ws.WriteMessage(websocket.TextMessage, []byte("Error: Username already in use."))
		mutex.Unlock()
		return
	}

	// Register the username and client
	usernames[username] = true
	clients[ws] = username

	// Notify users if Codegirl connects
	if username == "Codegirl" {
		broadcastAdminStatus(true)
	} else {
		// Send admin status to the newly connected user
		isCodegirlOnline := false
		for _, user := range clients {
			if user == "Codegirl" {
				isCodegirlOnline = true
				break
			}
		}
		statusMsg := Message{
			Type:    "admin_status",
			Content: fmt.Sprintf("Codegirl is %s.", map[bool]string{true: "online", false: "offline"}[isCodegirlOnline]),
		}
		err := ws.WriteJSON(statusMsg)
		if err != nil {
			fmt.Printf("Error sending admin status to %s: %v\n", username, err)
		}
	}
	mutex.Unlock()

	fmt.Printf("User connected: %s\n", username)

	// Listen for incoming messages
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("Error reading message from %s: %v\n", username, err)
			mutex.Lock()
			delete(clients, ws)
			delete(usernames, username)

			// Notify users if Codegirl disconnects
			if username == "Codegirl" {
				broadcastAdminStatus(false)
			}
			mutex.Unlock()
			break
		}

		// Handle different message types
		if msg.Type == "typing" {
			broadcastTypingEvent(msg.Sender, msg.Receiver)
		} else if msg.Type == "message" {
			msg.Sender = username
			broadcast <- msg
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client, username := range clients {
			// Send messages to the receiver
			if msg.Type == "message" && username == msg.Receiver {
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Printf("Error sending message to %s: %v\n", username, err)
					client.Close()
					mutex.Lock()
					delete(clients, client)
					mutex.Unlock()
				}
			}
		}
	}
}

func broadcastAdminStatus(isConnected bool) {
	// Notify all users about Codegirl's presence
	status := "offline"
	if isConnected {
		status = "online"
	}
	msg := Message{
		Type:    "admin_status",
		Content: fmt.Sprintf("Codegirl is %s.", status),
	}

	for client, username := range clients {
		// Only send to non-admin users
		if username != "Codegirl" {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Printf("Error notifying users about admin status: %v\n", err)
				client.Close()
				mutex.Lock()
				delete(clients, client)
				mutex.Unlock()
			}
		}
	}
}

func broadcastTypingEvent(sender, receiver string) {
	msg := Message{
		Type:     "typing",
		Sender:   sender,
		Receiver: receiver,
	}

	for client, username := range clients {
		if username == receiver {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Printf("Error sending typing event to %s: %v\n", username, err)
				client.Close()
				mutex.Lock()
				delete(clients, client)
				mutex.Unlock()
			}
		}
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Serve index.html for the root route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// WebSocket endpoint
	http.HandleFunc("/ws", handleConnections)

	// Start the message handler in a separate goroutine
	go handleMessages()

	// Start the server
	port := ":8080"
	fmt.Printf("Server started on http://localhost%s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
