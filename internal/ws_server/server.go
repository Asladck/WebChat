package ws_server

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"websckt/models"
)

const htmlDir = "./web/templates/html"

type WSServer interface {
	Start() error
	Stop() error
	Engine() *gin.Engine
}

type wsSrv struct {
	wsUpg     *websocket.Upgrader
	wsClients map[*websocket.Conn]struct{}
	mutex     *sync.RWMutex
	broadcast chan *models.WsMessage
	server    *http.Server
	engine    *gin.Engine
}

func NewWsServer(addr string) WSServer {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gin.Recovery(), gin.Logger())

	ws := &wsSrv{
		engine:    r,
		wsUpg:     &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		wsClients: make(map[*websocket.Conn]struct{}),
		mutex:     &sync.RWMutex{},
		broadcast: make(chan *models.WsMessage),
	}

	r.GET("/ws", ws.wsHandler)
	r.GET("/api/test", ws.testHandler)

	// 2. Статические файлы
	r.Static("/static", "./web/static")

	// 3. HTML шаблоны
	r.LoadHTMLFiles("./web/templates/index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	ws.server = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return ws
}
func (ws *wsSrv) Engine() *gin.Engine {
	return ws.engine
}

func (ws *wsSrv) Start() error {
	go ws.writeToClientsBroadcast()
	return ws.server.ListenAndServe()
}

func (ws *wsSrv) Stop() error {
	close(ws.broadcast)
	ws.mutex.Lock()
	for conn := range ws.wsClients {
		conn.Close()
		delete(ws.wsClients, conn)
	}
	ws.mutex.Unlock()
	return ws.server.Shutdown(nil)
}

func (ws *wsSrv) testHandler(c *gin.Context) {
	c.String(http.StatusOK, "Test is successful")
}

func (ws *wsSrv) wsHandler(c *gin.Context) {
	clientIP := c.Request.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		// Если заголовка нет, используем стандартный метод
		clientIP = strings.Split(c.Request.RemoteAddr, ":")[0]
	} else {
		// X-Forwarded-For может содержать цепочку IP (первый - исходный клиент)
		ips := strings.Split(clientIP, ",")
		clientIP = strings.TrimSpace(ips[0])
	}

	logrus.Infof("Real client IP: %s", clientIP)

	conn, err := ws.wsUpg.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("WebSocket upgrade error: %v", err)
		return
	}

	logrus.Infof("Client %s connected", conn.RemoteAddr().String())

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
		var msg models.WsMessage
		if err := conn.ReadJSON(&msg); err != nil {
			logrus.Errorf("WebSocket read error: %v", err)
			break
		}
		msg.IsMyMessage = false
		host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err == nil {
			msg.IPAddress = host
		}
		msg.Time = time.Now().Format("15:04")
		ws.broadcast <- &msg
	}
}

func (ws *wsSrv) writeToClientsBroadcast() {
	for msg := range ws.broadcast {
		ws.mutex.RLock()
		for client := range ws.wsClients {
			if msg.IsMyMessage && client.RemoteAddr().String() == msg.IPAddress {
				continue
			}
			if err := client.WriteJSON(msg); err != nil {
				logrus.Errorf("WebSocket write error: %v", err)
			}
		}
		ws.mutex.RUnlock()
	}
}
