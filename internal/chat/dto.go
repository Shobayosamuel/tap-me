package chat

type CreateRoomRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	IsPrivate bool `json:"is_private"`
}

type JoinRoomRequest struct {
	RoomID uint `json:"room_id" binding:"required"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}