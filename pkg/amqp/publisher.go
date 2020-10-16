package amqp

//go:generate mockery -name=Publisher
type Publisher interface {
	Publish(body []byte, contentType string) error
	Close() error
}
