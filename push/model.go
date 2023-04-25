package push

import (
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type WebPush struct {
	VapidKeys
}

type VapidKeys struct {
	VAPIDPublicKey  string `json:"vapidPublicKey"`
	VAPIDPrivateKey string `json:"vapidPrivateKey"`
}

type PushStatus int

const (
	PushStatusSuccess PushStatus = iota
	// a failure that may be resolved by retrying
	PushStatusTempFail
	// a failure for which a retry would not hekp
	PushStatusHardFail
)

type PushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
}

type Subscription struct {
	webpush.Subscription

	ID    string `json:"id"`
	Topic string `json:"topic"`
}

type NotificationOptions struct {
	TTL     int             `json:"ttl"`
	Urgency webpush.Urgency `json:"urgency"`
}

type Notification struct {
	Topic   string              `json:"topic" binding:"required"`
	ID      string              `json:"id" binding:"required"`
	Time    time.Time           `json:"time" binding:"required"`
	Payload PushPayload         `json:"payload" binding:"required"`
	Options NotificationOptions `json:"options"`
}
