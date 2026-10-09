package domain

import "time"

type Post struct {
	ID        int64     `json:"id"`
	AuthorID  int64     `json:"author_id"`
	Content   string    `json:"content"`
	Anonymous bool      `json:"anonymous"`
	Poll      *Poll     `json:"poll,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Author *User `json:"author,omitempty"`
}
