package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/destruc7i0n/webpush-api/push"
	"github.com/destruc7i0n/webpush-api/server"
	"github.com/destruc7i0n/webpush-api/store"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	store, err := store.NewStore()
	if err != nil {
		log.Fatal(err)
	}

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

	push := push.NewWebPush(vapidKeys.VAPIDPublicKey, vapidKeys.VAPIDPrivateKey)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	s := server.NewServer(":"+port, store, push)

	go func() {
		log.Println("[INFO] Starting API server...")
		err := s.Serve()
		if err != nil {
			log.Fatal("[ERROR] Server failed to start: ", err)
		}
	}()
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Shutdown(ctx)
	log.Println("[INFO] Server stopped")
}
