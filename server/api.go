package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/destruc7i0n/webpush-api/push"
	"github.com/destruc7i0n/webpush-api/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func (s *Server) newRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Length", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, newSuccessResponse("hello world"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/status", s.status)
		r.Get("/vapid", s.getVapidKey)

		r.Route("/topic/{id:[a-z0-9_-]+}", func(r chi.Router) {
			r.With(s.topicCtx).Get("/", s.getTopic)
			r.With(s.topicCtx).Post("/subscribe", s.subscribe)
			r.Post("/push", s.sendNotification)
		})
	})

	return r
}

func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) {
	topicId := chi.URLParam(r, "id")

	data := &subscriptionRequest{}
	if err := render.Bind(r, data); err != nil {
		render.JSON(w, r, newErrorResponse(fmt.Sprintf("failed to bind request: %v", err)))
		return
	}

	subscription := push.Subscription{
		Subscription: data.Subscription,
		Topic:        topicId,
		ID:           uuid.New().String(),
	}

	s.store.SetStruct(store.GetSubscriptionKey(topicId, subscription.ID), subscription)

	render.JSON(w, r, newSuccessResponse("subscription added"))
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	// get all the scheduled jobs
	schedulerJobs := s.scheduler.Jobs()

	jobs := make([]jobStatus, 0, len(schedulerJobs))
	for _, job := range schedulerJobs {
		jobs = append(jobs, jobStatus{
			Tags:      job.Tags(),
			StartTime: job.NextRun().Format(time.RFC3339),
		})
	}

	notifications, err := s.store.GetNotifications()
	if err != nil {
		render.JSON(w, r, newErrorResponse(fmt.Sprintf("failed to get notifications: %v", err)))
		return
	}

	subscriptions, err := s.store.GetSubscriptions("*")
	if err != nil {
		render.JSON(w, r, newErrorResponse(fmt.Sprintf("failed to get subscriptions: %v", err)))
		return
	}

	render.JSON(w, r, newStatusResponse(jobs, notifications, subscriptions))
}

func (s *Server) getVapidKey(w http.ResponseWriter, r *http.Request) {
	keys := s.push.GetVapidKeys()
	resp := newVapidKeyResponse(keys)
	render.JSON(w, r, resp)
}

func (s *Server) topicCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		topicId := chi.URLParam(r, "id")

		subscriptions, err := s.store.GetSubscriptions(topicId)
		if err != nil {
			render.JSON(w, r, newErrorResponse(fmt.Sprintf("failed to get subscriptions: %v", err)))
			return
		}

		ctx := context.WithValue(r.Context(), store.KeyTopic, subscriptions)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) getTopic(w http.ResponseWriter, r *http.Request) {
	topic := r.Context().Value(store.KeyTopic).([]push.Subscription)

	render.JSON(w, r, newTopicResponse(topic))
}

// func (s *Server) deleteTopic(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("Hello, World!"))
// }

func (s *Server) sendNotification(w http.ResponseWriter, r *http.Request) {
	topicId := chi.URLParam(r, "id")

	reqData := &notificationRequest{}
	if err := render.Bind(r, reqData); err != nil {
		render.JSON(w, r, newErrorResponse(fmt.Sprintf("failed to bind request: %v", err)))
		return
	}

	webPushPayload := push.PushPayload{
		Title: reqData.Title,
		Body:  reqData.Body,
		Icon:  reqData.Icon,
	}

	notificationTime := time.Time{} // zero time
	if reqData.Scheduled != "" {
		// parse utc time
		nt, err := time.Parse(time.RFC3339, reqData.Scheduled)
		if err != nil {
			render.JSON(w, r, newErrorResponse(fmt.Sprintf("failed to parse notification time: %v", err)))
			return
		}
		notificationTime = nt
	}

	instant := notificationTime.IsZero()

	s.ScheduleNotification(push.Notification{
		Topic:   topicId,
		ID:      uuid.New().String(),
		Time:    notificationTime,
		Payload: webPushPayload,
		Options: reqData.NotificationOptions,
	})

	if instant {
		render.JSON(w, r, newSuccessResponse("notification sent"))
	} else {
		render.JSON(w, r, newSuccessResponse("notification scheduled"))
	}
}
