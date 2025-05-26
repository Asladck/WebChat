package ws_server

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

const htmlDir = "./web/templates/html"

type WSServer interface {
	Start() error
	Stop() error
}
type wsSrv struct {
	mux       *http.ServeMux
	srv       *http.Server
	wsUpg     *websocket.Upgrader
	wsClients map[*websocket.Conn]struct{}
	mutex     *sync.RWMutex
	broadcast chan *wsMessage
}

func NewWsServer(addr string) WSServer {
	m := http.NewServeMux()
	wsSrc := &wsSrv{
		mux:       m,
		srv:       &http.Server{Addr: addr, Handler: m},
		wsUpg:     &websocket.Upgrader{},
		wsClients: map[*websocket.Conn]struct{}{},
		mutex:     &sync.RWMutex{},
		broadcast: make(chan *wsMessage),
	}
	return wsSrc
}
func (ws *wsSrv) Start() error {
	ws.mux.Handle("/", http.FileServer(http.Dir(htmlDir)))
	ws.mux.HandleFunc("/ws", ws.wsHandler)
	ws.mux.HandleFunc("/test", ws.testHandler)
	go ws.writeToClientsBroadcast()
	return ws.srv.ListenAndServe()
}
func (ws *wsSrv) testHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Test is successful"))
	if err != nil {
		log.Fatal("Test isn`t successful")
		return
	}
}
func (ws *wsSrv) Stop() error {
	close(ws.broadcast)
	ws.mutex.Lock()
	for conn := range ws.wsClients {
		conn.Close()
		delete(ws.wsClients, conn)
	}
	ws.mutex.Unlock()
	return ws.srv.Shutdown(context.Background())

}
func (ws *wsSrv) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.wsUpg.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("Error with ws connection %v", err)
		return
	}
	logrus.Infof("Client with %s ip is connected", conn.RemoteAddr().String())
	ws.mutex.Lock()
	ws.wsClients[conn] = struct{}{}
	ws.mutex.Unlock()
	go ws.readFromClient(conn)
}
func (ws *wsSrv) readFromClient(conn *websocket.Conn) {
	defer func() {
		conn.Close()
		ws.mutex.Lock()
		delete(ws.wsClients, conn)
		ws.mutex.Unlock()
	}()
	for {
		msg := new(wsMessage)
		err := conn.ReadJSON(msg)
		if err != nil {
			logrus.Errorf("Error with reading from WebSocket: %v", err)
			break
		}
		host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			logrus.Errorf("Error with : %v", err)
			return
		}
		msg.IPAddress = host
		msg.Time = time.Now().Format("15:04")
		ws.broadcast <- msg
	}
}
func (ws *wsSrv) writeToClientsBroadcast() {
	for msg := range ws.broadcast {
		ws.mutex.RLock()
		for client := range ws.wsClients {
			func() {
				if err := client.WriteJSON(msg); err != nil {
					logrus.Errorf("Error with writing message: %v", err)
				}
			}()
		}
		ws.mutex.RUnlock()
	}
}
