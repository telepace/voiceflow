package middleware

import (
	"github.com/gorilla/websocket"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func ErrorHandler(handler func(*websocket.Conn, interface{}) error) func(*websocket.Conn, interface{}) error {
	return func(conn *websocket.Conn, msg interface{}) error {
		if err := handler(conn, msg); err != nil {
			response := ErrorResponse{
				Error: err.Error(),
				Code:  500,
			}
			return conn.WriteJSON(response)
		}
		return nil
		
	}
}
