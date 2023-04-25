# webpush-api

A simple service for sending web push notifications to subscribers.

## API

### GET /api/vapid
**Response**
```json
{ "status": "success", "key": "..." }
```

### GET /api/status
**Response**
```json
{ "status": "...", "jobs": [...] ... }
```

### POST /api/topic/:topic
**Response**
```json
{ "status": "success", "subscriptions": [] }
```

### DELETE /api/topic/:topic
**Response**
```json
{ "status": "success" }
```

### POST /api/topic/:topic/subscribe
**Request Body**
```json
{ "subscription": { "endpoint": "...", "keys": { "p256dh": "...", "auth": "..." } } }
```

**Response**
```json
{ "status": "success", "id": "..." }
```

### POST /api/topic/:topic/push
**Request Body**
```json
{ "title": "...", "body": "...", "icon": "...", "scheduled": "...RFC 3339..." }
```
* Optional fields: `icon`, `scheduled`, `ttl`, `urgency`

**Response**
```json
{ "status": "success", "id": "...uuid..." }
```