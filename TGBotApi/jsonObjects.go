package TGBotApi

type Update struct {
	UpdateId          int     `json:"update_id"`
	Message           Message `json:"message, omitempty"`
	EditedMessage     Message `json:"edited_message, omitempty"`
	ChannelPost       Message `json:"channel_post, omitempty"`
	EditedChannelPost Message `json:"edited_channel_post, omitempty"`
}

type Message struct {
	MessageId int             `json:"message_id"`
	From      User            `json:"from, omitempty"`
	Date      int             `json:"date"`
	Chat      Chat            `json:"chat"`
	Text      string          `json:"text"`
	Entities  []MessageEntity `json:"entities,omitempty"`
	Photo     []PhotoSize     `json:"photo, omitempty"`
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type PhotoSize struct {
	FileId   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"file_size, omitempty"`
}

type User struct {
	Id           int `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	UserName     string `json:"user_name, omitempty"`
	LanguageCode string `json:"language_code, omitempty"`
}

type Chat struct {
	Id                          int64     `json:"id"`
	Type                        string    `json:"type"`
	Title                       string    `json:"title, omitempty"`
	Username                    string    `json:"username, omitempty"`
	FirstName                   string    `json:"first_name, omitempty"`
	LastName                    string    `json:"last_name, omitempty"`
	AllMembersAreAdministrators bool      `json:"all_members_are_administrators, omitempty"`
	Photo                       ChatPhoto `json:"photo, omitempty"`
	Description                 string    `json:"description, omitempty"`
	InviteLink                  string    `json:"invite_link, omitempty"`
}

type ChatPhoto struct {
	SmallFileId string `json:"small_file_id, omitempty"`
	BigFileId   string `json:"big_file_id, omitempty"`
}

type File struct {
	FileId   string `json:"file_id"`
	FileSize int    `json:"file_size, omitempty"`
	FilePath string `json:"file_path, omitempty"`
}
