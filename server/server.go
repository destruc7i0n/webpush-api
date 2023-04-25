package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/destruc7i0n/webpush-api/push"
	"github.com/destruc7i0n/webpush-api/store"
)

type Server struct {
	server    *http.Server
	store     *store.Store
	push      *push.WebPush
	scheduler *scheduler
	shutdown  bool
}

func NewServer(addr string, store *store.Store, push *push.WebPush) (s *Server) {
	scheduler := startScheduler()

	s = &Server{
		server:    nil,
		store:     store,
		push:      push,
		scheduler: scheduler,
	}

	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router(),
	}

	s.loadAndScheduleNotifications()

	return s
}

func (s *Server) loadAndScheduleNotifications() {
	notifications, err := s.store.GetAllNotifications()
	if err != nil {
		log.Printf("[ERROR] Failed to get notifications: %v", err)
		return
	}

	for _, notification := range notifications {
		s.ScheduleNotification(notification, false)
	}
}

func (s *Server) ScheduleNotification(notification push.Notification, instant bool) {
	s.store.SetStruct(store.GetNotificationKey(notification.ID), notification)

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
		s.store.Delete(store.GetNotificationKey(notification.ID))
	}

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

	return
}
