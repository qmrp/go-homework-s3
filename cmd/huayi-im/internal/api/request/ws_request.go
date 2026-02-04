package request

type WsRequest struct {
	Sid string `form:"sid" binding:"required"`
}

