package posts_transport_http

import (
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	polls_transport_http "github.com/oleg-morshel/murmur-api/internal/features/polls/transport/http"
)

type PostResponse struct {
	ID        int64                              `json:"id"`
	AuthorID  *int64                             `json:"author_id,omitempty"`
	Author    *string                            `json:"author,omitempty"`
	Content   string                             `json:"content"`
	Anonymous bool                               `json:"anonymous"`
	Poll      *polls_transport_http.PollResponse `json:"poll,omitempty"`
	CreatedAt time.Time                          `json:"created_at"`
	UpdatedAt time.Time                          `json:"updated_at"`
}

func NewPostResponse(post *domain.Post) PostResponse {
	resp := PostResponse{
		ID:        post.ID,
		Content:   post.Content,
		Anonymous: post.Anonymous,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	if !post.Anonymous {
		resp.AuthorID = &post.AuthorID
		if post.Author != nil {
			resp.Author = &post.Author.Username
		}
	}

	if post.Poll != nil {
		pollResp := polls_transport_http.NewPollResponse(post.Poll)
		resp.Poll = &pollResp
	}

	return resp
}

func NewPostListResponse(posts []*domain.Post) []PostResponse {
	result := make([]PostResponse, len(posts))
	for i, post := range posts {
		result[i] = NewPostResponse(post)
	}
	return result
}
