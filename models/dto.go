package models

type WsMessage struct {
	IPAddress   string `json:"address"`
	Username    string `json:"username"`
	Message     string `json:"message"`
	Time        string `json:"time"`
	IsMyMessage bool   `json:"isMyMessage"`
}
