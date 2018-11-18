package main

import (
	"github.com/dgrr/fastws"
)

// WebSocket is a wrapper around a gorilla.WebSocket for claws.
type WebSocket struct {
	conn      *fastws.Conn
	writeChan chan string
	closed    bool
	url       string
}

// URL returns the URL of the WebSocket.
func (w *WebSocket) URL() string {
	return w.url
}

// ReadChannel retrieves a channel from which to read messages out of.
func (w *WebSocket) ReadChannel() <-chan string {
	ch := make(chan string, 16)
	go w.readChannel(ch)
	return ch
}

// Write writes a message to the WebSocket
func (w *WebSocket) Write(msg string) {
	w.writeChan <- msg
}

func (w *WebSocket) readChannel(c chan<- string) {
	var _type fastws.Mode
	var msg []byte
	var err error
	for {
		_type, msg, err = w.conn.ReadMessage(msg[:0])
		if err != nil {
			if err == fastws.EOF {
				w.closed = true
			}
			if !w.closed {
				state.Error(err.Error())
				w.close(c)
			}
			return
		}

		switch _type {
		case fastws.ModeText, fastws.ModeBinary:
			c <- string(msg)
		}
	}
}

func (w *WebSocket) writePump() {
	for msg := range w.writeChan {
		_, err := w.conn.WriteMessage(fastws.ModeText, []byte(msg))
		if err != nil {
			state.Error(err.Error())
			w.Close()
			return
		}
	}
}

// Close closes the WebSocket connection.
func (w *WebSocket) Close() error {
	if w == nil {
		return nil
	}
	if w.closed {
		return nil
	}
	w.closed = true
	if state.Conn == w {
		state.Conn = nil
	}
	close(w.writeChan)
	return w.conn.Close("Bye bye :)")
}

// close finalises the WebSocket connection.
func (w *WebSocket) close(c chan<- string) error {
	close(c)
	return w.Close()
}

// WebSocketResponseError is the error returned when there is an error in
// CreateWebSocket.
type WebSocketResponseError struct {
	Err error
}

func (w WebSocketResponseError) Error() string {
	return w.Err.Error()
}

// CreateWebSocket initialises a new WebSocket connection.
func CreateWebSocket(url string) (*WebSocket, error) {
	state.Debug("Starting WebSocket connection to " + url)

	conn, err := fastws.Dial(url)
	if err != nil {
		return nil, WebSocketResponseError{
			Err: err,
		}
	}

	ws := &WebSocket{
		conn:      conn,
		writeChan: make(chan string, 128),
		url:       url,
	}

	go ws.writePump()

	return ws, nil
}
