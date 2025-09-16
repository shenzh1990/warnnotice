package util

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

type WSConn struct {
	conn   *websocket.Conn
	Closed bool
	Url    string
	WSChan chan []byte
}

func NewWSConn(url string) *WSConn {
	dialer := &websocket.Dialer{
		Proxy:          http.ProxyFromEnvironment,
		ReadBufferSize: 1024 * 32,
	}
	c, _, connErr := dialer.Dial(url, nil)
	if connErr != nil {
		Log(LogLevelError, fmt.Sprintf("WebSocket connection failed:%v", connErr))
	}
	return &WSConn{
		conn:   c,
		Closed: false,
		Url:    url,
		WSChan: make(chan []byte, 50),
	}
}
func (wsConn *WSConn) Handle() {
	for {
		msg := <-wsConn.WSChan
		if len(wsConn.WSChan) > 10 {
			Log(LogLevelError, fmt.Sprintf(`wsConn wait list 10 %d %s %#v`, len(wsConn.WSChan), string(msg), wsConn))
			continue
		}
		var err error
		err = wsConn.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			wsConn.Closed = true
			Log(LogLevelError, `handle ws err `+err.Error())
		}
	}
}
func (wsConn *WSConn) WriteJson(msg interface{}) (err error) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return wsConn.WriteMessage(jsonData)
}
func (wsConn *WSConn) WriteMessage(msg []byte) (err error) {
	if wsConn.conn == nil || wsConn.Closed {
		return fmt.Errorf(`nil conn`)
	}
	wsConn.WSChan <- msg
	return
}
func (wsConn *WSConn) ReadMessage() (messageType int, p []byte, err error) {
	return wsConn.conn.ReadMessage()
}
func (wsConn *WSConn) Close() error {
	wsConn.Closed = true
	err := wsConn.conn.Close()
	if err != nil {
		Log(LogLevelError, `close conn err `+err.Error())
		return err
	}
	return nil
}
