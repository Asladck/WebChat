package models

type WsMessage struct {
	IPAddress   string `json:"address"`
	Message     string `json:"message"`
	Time        string `json:"time"`
	IsMyMessage bool   `json:"isMyMessage"`
}
