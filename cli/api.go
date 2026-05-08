package cli
import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"

	"coop-cloud-backend/internal"
)
// Upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
       return true
    },
}
func wsHandler(w http.ResponseWriter, r *http.Request) {
    // Upgrade the HTTP connection to a WebSocket connection
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
       fmt.Println("Error upgrading:", err)
       return
    }
    defer conn.Close()
    // Listen for incoming messages
    for {
       // Read message from the client
       _, message, err := conn.ReadMessage()
       if err != nil {
          fmt.Println("Error reading message:", err)
          break
       }
       fmt.Printf("Received: %s\\n", message)
       // Echo the message back to the client
       if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
          fmt.Println("Error writing message:", err)
          break
       }
    }
}

type abraHandler struct{
	mux *http.ServeMux
}
func newAbraHandler() *abraHandler {
	h := &abraHandler{
		mux: http.NewServeMux(),
	}
	h.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API Path not found")
		http.Error(w, "API Path not found", http.StatusMethodNotAllowed)
	})
	h.mux.HandleFunc("/api/apps", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet: 
			h.handleListApps(w, r)
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/apps/{appId}/deploy", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Chaos") == "true"{
			internal.Chaos = true
		} else {
			internal.Chaos = false
		}
			
		switch r.Method{
		case http.MethodPost:
			h.handleDeployApp(w, r, r.PathValue("appId"))
		case http.MethodGet:
			h.getDeployLogs(w, r, r.PathValue("appId"))
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/apps/{appId}/stop", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method{
		case http.MethodPost:
			h.handleUndeployApp(w, r, r.PathValue("appId"))
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/apps/{appId}/remove", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method{
		case http.MethodPost:
			h.handleRemoveApp(w, r, r.PathValue("appId"))
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/servers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet: 
			h.handleListServers(w, r)
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/catalogue", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet: 
			h.handleListCatalogue(w, r)
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})

	h.mux.HandleFunc("/api/apps/{appId}/new", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method{
		case http.MethodPost:
			h.handleNewApp(w, r, r.PathValue("appId"))
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/apps/{appId}/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method{
		case http.MethodPost:
			h.handlePostConfig(w, r, r.PathValue("appId"))
		case http.MethodGet:
			h.handleGetConfig(w, r, r.PathValue("appId"))
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/apps/{appId}/secret", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method{
		case http.MethodPost:
			h.handleInsertAppSecret(w, r, r.PathValue("appId"))
		case http.MethodGet:
			h.handleGetAppSecrets(w, r, r.PathValue("appId"))
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/apps/{appId}/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method{
		case http.MethodGet:
			h.handleGetAppServices(w, r, r.PathValue("appId"))
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	h.mux.HandleFunc("/api/apps/{appId}/{serviceId}/logs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method{
		case http.MethodGet:
			h.handleGetLogs(w, r, r.PathValue("appId"), r.PathValue("serviceId"))
		default:
			http.Error(w, "Method not implemented", http.StatusMethodNotAllowed)
		}
	})
	return h
}
func StartAPI() {
	/*http.HandleFunc("/ws", wsHandler)
    fmt.Println("WebSocket server started on :3001")
	err := http.ListenAndServe(":3001", nil)
    if err != nil {
       fmt.Println("Error starting server:", err)
    }*/

	h := newAbraHandler()
	fmt.Println("Server started on port 3000")
	http.ListenAndServe(":3000", h)
}

func (h *abraHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Pre-processing: logging
	log.Printf("Incoming %s request: %s\n", r.Method, r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	// Delegate to internal mux
	h.mux.ServeHTTP(w, r)
}
