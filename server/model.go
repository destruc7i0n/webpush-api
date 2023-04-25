package server

import (
	"net/http"

	"github.com/destruc7i0n/webpush-api/push"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// requests

type subscriptionRequest struct {
	Subscription webpush.Subscription `json:"subscription"`
}

func (sr *subscriptionRequest) Bind(r *http.Request) error {
	return nil
}

type notificationRequest struct {
	push.PushPayload
	push.NotificationOptions

	Scheduled string `json:"scheduled,omitempty"`
}

func (nr *notificationRequest) Bind(r *http.Request) error {
	if nr.Urgency == "" {
		nr.Urgency = webpush.UrgencyNormal
	}
	return nil
}

// responses

type ResponseType string

const (
	ResponseTypeSuccess ResponseType = "success"
	ResponseTypeError   ResponseType = "error"
)

type response struct {
	Status  ResponseType `json:"status"`
	Message string       `json:"message,omitempty"`
}

func newSuccessResponse(message string) *response {
	return &response{
		Status:  ResponseTypeSuccess,
		Message: message,
	}
}

type errorResponse struct {
	response
}

func newErrorResponse(message string) *errorResponse {
	return &errorResponse{
		response: response{
			Status:  ResponseTypeError,
			Message: message,
		},
	}
}

type vapidKeyResponse struct {
	response
	Key string `json:"key"`
}

func newVapidKeyResponse(vapidKeys push.VapidKeys) *vapidKeyResponse {
	return &vapidKeyResponse{
		response: response{
			Status: ResponseTypeSuccess,
		},
		Key: vapidKeys.VAPIDPublicKey,
	}
}

type topicResponse struct {
	response
	Topic []push.Subscription `json:"topic"`
}

func newTopicResponse(topic []push.Subscription) *topicResponse {
	return &topicResponse{
		response: response{
			Status: ResponseTypeSuccess,
		},
		Topic: topic,
	}
}

type jobStatus struct {
	Tags      []string `json:"tags"`
	StartTime string   `json:"startTime"`
}

type statusResponse struct {
	response
	Jobs          []jobStatus         `json:"jobs"`
	Notifications []push.Notification `json:"notifications"`
	Subscriptions []push.Subscription `json:"subscriptions"`
}

func newStatusResponse(jobs []jobStatus, notifications []push.Notification, subscriptions []push.Subscription) *statusResponse {
	return &statusResponse{
		response: response{
			Status: ResponseTypeSuccess,
		},
		Jobs:          jobs,
		Notifications: notifications,
		Subscriptions: subscriptions,
	}
}
