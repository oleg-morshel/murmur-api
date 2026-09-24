package polls_transport_http

import (
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

type PollOptionResponse struct {
	ID    int64  `json:"id"`
	Text  string `json:"text"`
	Votes int    `json:"votes"`
}

type PollResponse struct {
	ID        int64                `json:"id"`
	PostID    int64                `json:"post_id"`
	Question  string               `json:"question"`
	Options   []PollOptionResponse `json:"options"`
	CreatedAt time.Time            `json:"created_at"`
}

func NewPollResponse(poll *domain.Poll) PollResponse {
	options := make([]PollOptionResponse, 0, len(poll.Options))
	for _, opt := range poll.Options {
		options = append(options, PollOptionResponse{
			ID:    opt.ID,
			Text:  opt.Text,
			Votes: opt.Votes,
		})
	}

	return PollResponse{
		ID:        poll.ID,
		PostID:    poll.PostID,
		Question:  poll.Question,
		Options:   options,
		CreatedAt: poll.CreatedAt,
	}
}
