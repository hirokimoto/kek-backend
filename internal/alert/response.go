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
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Account   Account   `json:"account"`
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
	Account   Account   `json:"account"`
}

type Account struct {
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
	return &AlertResponse{
		Alert: Alert{
			Slug:      a.Slug,
			Title:     a.Title,
			Body:      a.Body,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			Account: Account{
				Username: a.Account.Username,
				Bio:      a.Account.Bio,
				Image:    a.Account.Image,
			},
		},
	}
}
