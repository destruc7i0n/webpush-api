package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/destruc7i0n/webpush-api/push"
	"github.com/destruc7i0n/webpush-api/store"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type Server struct {
	server    *http.Server
	store     *store.Store
	push      *push.WebPush
	scheduler *scheduler
	shutdown  bool
	notifs    chan *push.Notification
}

func NewServer(addr string, store *store.Store) (s *Server) {
	// init vapid keys
	vapidKeys, err := store.GetVapidKeys()
	if err != nil {
		vapidKeys = push.GenerateVAPIDKeys()
		err = store.SetVapidKeys(vapidKeys)
		if err != nil {
			log.Fatal("[ERROR] Failed to set VAPID keys: ", err)
		}
		log.Printf("[INFO] Generated VAPID keys: %+v", vapidKeys)
	}

	// init webpush
	wp := push.NewWebPush(vapidKeys.VAPIDPublicKey, vapidKeys.VAPIDPrivateKey)

	// init scheduler
	scheduler := startScheduler()

	s = &Server{
		server:    nil,
		store:     store,
		push:      wp,
		scheduler: scheduler,
		shutdown:  false,
		notifs:    make(chan *push.Notification, 256),
	}

	s.server = &http.Server{
		Addr:    addr,
		Handler: s.newRouter(),
	}

	s.loadAndScheduleNotifications()
	go s.startNotificationChannel()

	return s
}

func (s *Server) loadAndScheduleNotifications() {
	notifications, err := s.store.GetNotifications()
	if err != nil {
		log.Printf("[ERROR] Failed to get notifications: %v", err)
		return
	}

	for _, notification := range notifications {
		s.ScheduleNotification(notification)
	}
}

func (s *Server) startNotificationChannel() {
	for {
		notification := <-s.notifs
		s.ScheduleNotification(*notification)
	}
}

func (s *Server) ScheduleNotification(notification push.Notification) {
	notificationKey := store.GetNotificationKey(notification.Topic, notification.ID)
	s.store.SetStruct(notificationKey, notification)

	job := func() {
		subscriptions, err := s.store.GetSubscriptions(notification.Topic)

		if err != nil {
			log.Printf("[ERROR] Failed to get subscriptions: %v", err)
			return
		}

		options := webpush.Options{
			Topic:   notification.Topic,
			TTL:     notification.Options.TTL,
			Urgency: notification.Options.Urgency,
		}

		for _, subscription := range subscriptions {
			status := s.push.Send(&subscription.Subscription, notification.Payload, options)
			if status != push.PushStatusSuccess {
				log.Printf("[ERROR] Failed to send notification: %v", status)

				if status == push.PushStatusHardFail {
					// if fail, delete subscription
					s.store.Delete(store.GetSubscriptionKey(notification.Topic, subscription.ID))
				}
			}
		}

		// delete notification from store
		s.store.Delete(notificationKey)
	}

	instant := notification.Time.IsZero()

	// check if in the past
	if !instant && notification.Time.Before(time.Now()) {
		log.Printf("[INFO] Notification %s is in the past, sending now", notification.ID)
		instant = true
	}

	if instant {
		log.Printf("[INFO] Sending notification %s now", notification.ID)
		job()
	} else {
		log.Printf("[INFO] Scheduling notification %s at %s", notification.ID, notification.Time)
		s.scheduler.scheduleAt(notification.Time, notification.Topic, job)
	}
}

func (s *Server) Serve() (err error) {
	log.Printf("[INFO] Server is listening on %s", s.server.Addr)
	err = s.server.ListenAndServe()
	if err != nil && !s.shutdown {
		log.Printf("[ERROR] Server failed to start: %v", err)
		return err
	}
	err = nil
	return
}

func (s *Server) Shutdown(ctx context.Context) (err error) {
	s.shutdown = true

	if err = s.store.Close(); err != nil {
		log.Printf("[ERROR] Store close error: %v", err)
		return err
	}

	if err = s.server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Server shutdown error: %v", err)
		return err
	}

	s.scheduler.Stop()

	return
}
