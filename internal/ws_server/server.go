package ws_server

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
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
	broadcast chan *BroadcastPayload
	server    *http.Server
	engine    *gin.Engine
}
type BroadcastPayload struct {
	Msg    *models.WsMessage
	Sender *websocket.Conn
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
		broadcast: make(chan *BroadcastPayload),
	}
	r.LoadHTMLFiles("./web/templates/index.html", "./web/templates/chat.html", "./web/templates/register.html", "./web/templates/login.html")
	r.GET("/sign-up", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.GET("/sign-in", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/chat", func(c *gin.Context) {
		c.HTML(http.StatusOK, "chat.html", nil)
	})
	r.GET("/ws", ws.wsHandler)
	r.GET("/api/test", ws.testHandler)

	// 2. Статические файлы
	r.Static("/static", "./web/static")

	// 3. HTML шаблоны
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
	// 1. Получение и проверка JWT-токена из query
	tokenStr := c.Query("token")
	if tokenStr == "" {
		logrus.Warn("WebSocket connection attempt without token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Проверка алгоритма подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("qweqroqwro123e21edwqdl@@"), nil
	})

	if err != nil || !token.Valid {
		logrus.WithError(err).Warn("Invalid JWT token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 2. Извлечение username из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logrus.Warn("Failed to parse JWT claims")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		logrus.Warn("Username not found in token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 3. Получение IP-адреса клиента
	clientIP := getClientIP(c.Request)
	logrus.WithFields(logrus.Fields{
		"username": username,
		"ip":       clientIP,
		"agent":    c.Request.UserAgent(),
	}).Info("New WebSocket connection")

	// 4. Апгрейд соединения до WebSocket
	conn, err := ws.wsUpg.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.WithError(err).Error("WebSocket upgrade failed")
		return
	}

	// 5. Регистрация клиента
	ws.mutex.Lock()
	ws.wsClients[conn] = struct{}{}
	ws.mutex.Unlock()

	// 6. Запуск обработчика входящих сообщений
	go ws.readFromClient(conn, username)
}
func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Берём первый IP из цепочки
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func (ws *wsSrv) readFromClient(conn *websocket.Conn, username string) {
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
		msg.Username = username
		msg.Time = time.Now().Format("15:04")
		host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err == nil {
			msg.IPAddress = host
		}
		ws.broadcast <- &BroadcastPayload{
			Msg:    &msg,
			Sender: conn,
		}
	}
}

func (ws *wsSrv) writeToClientsBroadcast() {
	for payload := range ws.broadcast {
		ws.mutex.RLock()
		for client := range ws.wsClients {
			if client == payload.Sender {
				continue
			}
			if err := client.WriteJSON(payload.Msg); err != nil {
				logrus.Errorf("WebSocket write error: %v", err)
			}
		}
		ws.mutex.RUnlock()
	}
}
