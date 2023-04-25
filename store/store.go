package store

import (
	"encoding/json"
	"fmt"

	"github.com/destruc7i0n/webpush-api/push"
)

type StoreKey string

const (
	KeyVapidKeys    StoreKey = "vapidKeys"
	KeyTopic        StoreKey = "topic"
	KeySubscription StoreKey = "subscription"
	KeyNotification StoreKey = "notification"
)

func GetTopicKey(topic string) string {
	return fmt.Sprintf("%s:%s", KeyTopic, topic)
}

func GetSubscriptionKey(topic, id string) string {
	return fmt.Sprintf("%s:%s:%s", GetTopicKey(topic), KeySubscription, id)
}

func GetNotificationKey(id string) string {
	return fmt.Sprintf("%s:%s", KeyNotification, id)
}

func (s *Store) GetVapidKeys() (push.VapidKeys, error) {
	var vapidKeys push.VapidKeys
	err := s.GetStruct(string(KeyVapidKeys), &vapidKeys)
	return vapidKeys, err
}

func (s *Store) SetVapidKeys(vapidKeys push.VapidKeys) error {
	return s.SetStruct(string(KeyVapidKeys), vapidKeys)
}

func (s *Store) GetSubscriptions(topic string) ([]push.Subscription, error) {
	subs, err := s.ListByPrefix(GetSubscriptionKey(topic, ""))
	if err != nil {
		return nil, err
	}

	subscriptions := make([]push.Subscription, 0, len(subs))
	for _, sub := range subs {
		var subscription push.Subscription
		if err := json.Unmarshal([]byte(sub), &subscription); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

func (s *Store) AddNotification(topic string, notification push.Notification) error {
	return s.SetStruct(GetNotificationKey(notification.ID), notification)
}

func (s *Store) GetAllNotifications() ([]push.Notification, error) {
	notifications, err := s.ListByPrefix(GetNotificationKey(""))
	if err != nil {
		return nil, err
	}

	resp := make([]push.Notification, 0, len(notifications))
	for _, notification := range notifications {
		var notif push.Notification
		if err := json.Unmarshal([]byte(notification), &notif); err != nil {
			return nil, err
		}
		resp = append(resp, notif)
	}

	return resp, nil
}
