package push

import (
	"encoding/json"
	"io"
	"log"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func GenerateVAPIDKeys() VapidKeys {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatal("[ERROR] Failed to generate VAPID keys: ", err)
	}

	return VapidKeys{publicKey, privateKey}
}

func NewWebPush(vapidPublicKey, vapidPrivateKey string) (wp *WebPush) {
	wp = &WebPush{VapidKeys{vapidPublicKey, vapidPrivateKey}}
	return
}

func (w *WebPush) GetVapidKeys() VapidKeys {
	return w.VapidKeys
}

func (w *WebPush) Send(subscription *webpush.Subscription, payload *PushPayload, options *webpush.Options) PushStatus {
	p, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal payload: %s", err)
		return PushStatusHardFail
	}

	// combine options
	if options == nil {
		options = &webpush.Options{}
	}
	options.Subscriber = "mail@thedestruc7i0n.ca"
	options.VAPIDPublicKey = w.VAPIDPublicKey
	options.VAPIDPrivateKey = w.VAPIDPrivateKey

	// log.Printf("[INFO] Sending push with options: %+v", options)

	startedAt := time.Now()
	resp, err := webpush.SendNotification(p, subscription, options)

	if err != nil {
		log.Printf("[ERROR] Failed to push: %s", err)
		return PushStatusHardFail
	}

	defer resp.Body.Close()
	duration := time.Since(startedAt)
	log.Printf("[INFO] Pushed (%d) in %s", resp.StatusCode, duration.String())

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case 201:
		log.Println("[INFO] Push accepted by push service")
		return PushStatusSuccess
	case 429:
		log.Println("[INFO] Push rejected by push service (rate limit)", body)
		return PushStatusTempFail
	case 400, 404, 405, 413, 500, 501:
		// Bad Request, Not Found, Method Not Allowed, Payload Too Large, Internal Server Error, Not Implemented
		log.Println("[INFO] Push rejected by push service:", resp.StatusCode)
	case 410:
		log.Println("[INFO] Subscription expired")
	default:
		log.Printf("[INFO] Unknown response from push service: %d", resp.StatusCode)
	}

	log.Printf("[INFO] Push failed. Body: %s", body)

	return PushStatusHardFail
}
