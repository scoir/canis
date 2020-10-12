package notifier

const QueueName = "notification"

type Notification struct {
	Topic     string      `json:"topic"`
	Event     string      `json:"event"`
	EventData interface{} `json:"message"`
}

type EventMessage struct {
	Event     string      `json:"event"`
	Timestamp int64       `json:"timestamp"`
	EventData interface{} `json:"message"`
}
