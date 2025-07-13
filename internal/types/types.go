package types

// GRUniChat消息结构体
type GRUniChatMessage struct {
	From        string                 `json:"from"`
	Type        string                 `json:"type"`
	Body        GRUniChatBody          `json:"body"`
	TotalID     string                 `json:"totalId"`
	CurrentTime string                 `json:"currentTime"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

type GRUniChatBody struct {
	Sender      string `json:"sender"`
	ChatMessage string `json:"chatMessage"`
	Command     string `json:"command,omitempty"`
	EventDetail string `json:"eventDetail,omitempty"`
	ExecuteAt   string `json:"executeAt,omitempty"`
}

// OneBot v11消息结构体
type OneBotMessage struct {
	PostType    string                 `json:"post_type"`
	MessageType string                 `json:"message_type,omitempty"`
	SubType     string                 `json:"sub_type,omitempty"`
	MessageID   int64                  `json:"message_id,omitempty"`
	UserID      int64                  `json:"user_id,omitempty"`
	GroupID     int64                  `json:"group_id,omitempty"`
	Message     interface{}            `json:"message,omitempty"` // 可能是string或array
	RawMessage  string                 `json:"raw_message,omitempty"`
	Font        int                    `json:"font,omitempty"`
	Sender      OneBotSender           `json:"sender,omitempty"`
	Time        int64                  `json:"time,omitempty"`
	SelfID      int64                  `json:"self_id,omitempty"`
	Extra       map[string]interface{} `json:"-"`
}

type OneBotSender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card,omitempty"`
	Role     string `json:"role,omitempty"`
	Title    string `json:"title,omitempty"`
}

// OneBot消息段结构体
type MessageSegment struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// 待确认的命令结构体
type PendingCommand struct {
	UserID      int64  `json:"user_id"`
	GroupID     int64  `json:"group_id"`
	Command     string `json:"command"`
	OriginalMsg string `json:"original_msg"`
	Timestamp   int64  `json:"timestamp"`
}

// OneBot API响应结构体
type OneBotResponse struct {
	Status  string      `json:"status"`
	RetCode int         `json:"retcode"`
	Data    interface{} `json:"data,omitempty"`
	Echo    string      `json:"echo,omitempty"`
}
