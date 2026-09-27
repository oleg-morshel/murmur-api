package core_events

import "time"

const (
	SubjectPostCreated = "post.created"
	SubjectPostDeleted = "post.deleted"
	SubjectPollVoted   = "poll.voted"
)

type PostCreatedEvent struct {
	Type      string    `json:"type"`
	PostID    int64     `json:"post_id"`
	AuthorID  int64     `json:"author_id"`
	Timestamp time.Time `json:"timestamp"`
}

type PostDeletedEvent struct {
	Type      string    `json:"type"`
	PostID    int64     `json:"post_id"`
	Timestamp time.Time `json:"timestamp"`
}

type PollVotedEvent struct {
	Type      string    `json:"type"`
	PollID    int64     `json:"poll_id"`
	OptionID  int64     `json:"option_id"`
	UserID    int64     `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}
