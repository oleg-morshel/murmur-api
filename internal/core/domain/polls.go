package domain

import "time"

type Poll struct {
	ID        int64         `json:"id"`
	PostID    int64         `json:"post_id"`
	Question  string        `json:"question"`
	Options   []*PollOption `json:"options"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type PollOption struct {
	ID       int64  `json:"id"`
	PollID   int64  `json:"poll_id"`
	Text     string `json:"text"`
	Votes    int    `json:"votes"`
	Position int    `json:"position"`
}

type PollVote struct {
	ID        int64     `json:"id"`
	PollID    int64     `json:"poll_id"`
	OptionID  int64     `json:"option_id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
