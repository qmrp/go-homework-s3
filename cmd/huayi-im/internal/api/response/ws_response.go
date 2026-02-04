package response

/*
*
* @Description: 响应结构- 心跳包响应
**/
type PongResponse struct {
	MessageType string `json:"message-type"`
	To          string `json:"to"`
	MessageID   int64  `json:"message-id"`
}

// WebSocketMessage WebSocket消息结构
type WebSocketMessage struct {
	MessageID   int64    `json:"message-id"`
	From        string   `json:"from"`
	To          []string `json:"to,omitempty"`
	Topic       string   `json:"topic,omitempty"`
	MessageType string   `json:"message-type"`
	ContentType string   `json:"content-type,omitempty"`
	Content     string   `json:"content,omitempty"`
	AckID       int64    `json:"ack-id,omitempty"`
}
