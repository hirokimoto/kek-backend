package alert

import (
	"kek-backend/internal/alert/model"
	"time"
)

type AlertResponse struct {
	Alert Alert `json:"alert"`
}

type AlertsResponse struct {
	Alert       []Alert `json:"alerts"`
	AlertsCount int64   `json:"alertsCount"`
}

type Alert struct {
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Tags      []string  `json:"tagList"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Author    Author    `json:"author"`
}

type CommentResponse struct {
	Comment Comment `json:"comment"`
}

type CommentsResponse struct {
	Comments []Comment `json:"comments"`
}

type Comment struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Body      string    `json:"body"`
	Author    Author    `json:"author"`
}

type Author struct {
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

// NewAlertsResponse converts alert models and total count to AlertsResponse
func NewAlertsResponse(alerts []*model.Alert, total int64) *AlertsResponse {
	var a []Alert
	for _, alert := range alerts {
		a = append(a, NewAlertResponse(alert).Alert)
	}

	return &AlertsResponse{
		Alert:       a,
		AlertsCount: total,
	}
}

// NewAlertResponse converts alert model to AlertResponse
func NewAlertResponse(a *model.Alert) *AlertResponse {
	var tags []string
	for _, tag := range a.Tags {
		tags = append(tags, tag.Name)
	}

	return &AlertResponse{
		Alert: Alert{
			Slug:      a.Slug,
			Title:     a.Title,
			Body:      a.Body,
			Tags:      tags,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			Author: Author{
				Username: a.Author.Username,
				Bio:      a.Author.Bio,
				Image:    a.Author.Image,
			},
		},
	}
}

// NewCommentsResponse converts alert comment models to CommentsResponse
func NewCommentsResponse(comments []*model.Comment) *CommentsResponse {
	var commentsRes []Comment
	for _, comment := range comments {
		commentsRes = append(commentsRes, NewCommentResponse(comment).Comment)
	}
	return &CommentsResponse{
		Comments: commentsRes,
	}
}

// NewCommentResponse converts alert comment model to CommentResponse
func NewCommentResponse(comment *model.Comment) *CommentResponse {
	return &CommentResponse{
		Comment: Comment{
			ID:        comment.ID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
			Body:      comment.Body,
			Author: Author{
				Username: comment.Author.Username,
				Bio:      comment.Author.Bio,
				Image:    comment.Author.Image,
			},
		},
	}
}
