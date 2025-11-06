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
type tokenClaims struct {
	jwt.StandardClaims
	UserId   string `json:"id"`
	Username string `json:"username"`
}

func InitRoutes(ws *wsSrv, r *gin.Engine) *gin.Engine {
	r.LoadHTMLFiles("./web/templates/profile.html", "./web/templates/index.html", "./web/templates/chat.html", "./web/templates/register.html", "./web/templates/login.html")
	r.GET("/sign-up", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.GET("/sign-in", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/chat", func(c *gin.Context) {
		c.HTML(http.StatusOK, "chat.html", nil)
	})
	r.GET("/profile/:username", func(c *gin.Context) {
		username := c.Param("username")
		c.HTML(http.StatusOK, "profile.html", gin.H{"username": username})
	})
	r.GET("/ws", ws.wsHandler)
	r.GET("/api/test", ws.testHandler)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.Static("/static", "./web/static")
	return r
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

	r = InitRoutes(ws, r)

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
	tokenStr := c.Query("token")
	if tokenStr == "" {
		logrus.Warn("WebSocket connection attempt without token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else {
		logrus.Println("Token is : " + tokenStr)
	}

	token, err := jwt.ParseWithClaims(tokenStr, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
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

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || claims.Username == "" {
		logrus.Warn("Username not found in token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	username := claims.Username

	logrus.Println("username: " + username)
	clientIP := getClientIP(c.Request)
	logrus.WithFields(logrus.Fields{
		"ip":    clientIP,
		"agent": c.Request.UserAgent(),
	}).Info("New WebSocket connection")

	conn, err := ws.wsUpg.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.WithError(err).Error("WebSocket upgrade failed")
		return
	}

	ws.mutex.Lock()
	ws.wsClients[conn] = struct{}{}
	ws.mutex.Unlock()

	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	go func(cConn *websocket.Conn) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := cConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					logrus.WithError(err).Info("Ping failed, closing connection")
					cConn.Close()
					ws.mutex.Lock()
					delete(ws.wsClients, cConn)
					ws.mutex.Unlock()
					return
				}
			}
		}
	}(conn)

	go ws.readFromClient(conn, username)
}
func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
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
				ws.mutex.RUnlock()
				ws.mutex.Lock()
				client.Close()
				delete(ws.wsClients, client)
				ws.mutex.Unlock()
				ws.mutex.RLock()
			}
		}
		ws.mutex.RUnlock()
	}
}
