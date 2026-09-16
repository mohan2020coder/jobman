package realtime

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jobman/backend/internal/auth"
	"github.com/jobman/backend/pkg/httpapi"
	"github.com/jobman/backend/internal/technicians"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// The mobile app sends no Origin header; the web dashboards are served
	// same-host. Allow all origins at the transport level and rely on the
	// signed token for authorization.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ServeWS authenticates the ?token= query param via the shared JWT secret and,
// for TECHNICIAN users, subscribes their socket to live job events.
func ServeWS(hub *Hub, techSvc *technicians.Service, secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.URL.Query().Get("token")
		if raw == "" {
			httpapi.WriteError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Missing token.")
			return
		}
		claims, err := auth.Parse(secret, raw)
		if err != nil {
			httpapi.WriteError(w, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "Session expired, please log in again.")
			return
		}
		if claims.BusinessID == "" || claims.UserID == "" || claims.Role != "TECHNICIAN" {
			httpapi.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Only technician accounts receive job events.")
			return
		}

		tech, err := techSvc.GetByUserID(r.Context(), claims.BusinessID, claims.UserID)
		if err != nil {
			httpapi.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Technician profile not found.")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("realtime: upgrade: %v", err)
			return
		}

		client := &Client{conn: conn, techID: tech.ID, send: make(chan []byte, sendBuffer)}
		hub.register(tech.ID, client)
		log.Printf("realtime: technician %s connected", tech.ID)
		client.start()
	}
}

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}