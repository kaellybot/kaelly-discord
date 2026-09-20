package testutils

import (
	"sync"

	amqp "github.com/kaellybot/kaelly-amqp"
)

// BrokerCall is one Emit, Request or Reply observed by FakeBroker.
type BrokerCall struct {
	Method        string
	Message       *amqp.RabbitMQMessage
	Exchange      amqp.Exchange
	RoutingKey    string
	CorrelationID string
	ReplyTo       string
}

// FakeBroker is a recording amqp.MessageBroker. Set Err to make every
// publishing call fail. Consume only stores the consumer so tests can drive it.
type FakeBroker struct {
	Err       error
	Connected bool

	// OnRequest, when set, runs inside Request before it returns, so a test can
	// deliver an answer that races the caller registering it. Set it before the
	// first call.
	OnRequest func(correlationID string)

	mu        sync.Mutex
	calls     []BrokerCall
	consumers map[string]amqp.MessageConsumer
}

var _ amqp.MessageBroker = (*FakeBroker)(nil)

func (b *FakeBroker) record(call BrokerCall) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls = append(b.calls, call)
	return b.Err
}

func (b *FakeBroker) Run() error { return nil }

func (b *FakeBroker) Emit(msg *amqp.RabbitMQMessage, exchange amqp.Exchange,
	routingKey, correlationID string) error {
	return b.record(BrokerCall{Method: "Emit", Message: msg, Exchange: exchange,
		RoutingKey: routingKey, CorrelationID: correlationID})
}

func (b *FakeBroker) Request(msg *amqp.RabbitMQMessage, exchange amqp.Exchange,
	routingKey, correlationID, replyTo string) error {
	err := b.record(BrokerCall{Method: "Request", Message: msg, Exchange: exchange,
		RoutingKey: routingKey, CorrelationID: correlationID, ReplyTo: replyTo})
	if b.OnRequest != nil {
		b.OnRequest(correlationID)
	}
	return err
}

func (b *FakeBroker) Reply(msg *amqp.RabbitMQMessage, correlationID, replyTo string) error {
	return b.record(BrokerCall{Method: "Reply", Message: msg,
		CorrelationID: correlationID, ReplyTo: replyTo})
}

func (b *FakeBroker) Consume(queueName string, consumer amqp.MessageConsumer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.consumers == nil {
		b.consumers = make(map[string]amqp.MessageConsumer)
	}
	b.consumers[queueName] = consumer
}

func (b *FakeBroker) IsConnected() bool { return b.Connected }
func (b *FakeBroker) Shutdown()         {}

// Calls returns a snapshot of every publishing call, in order.
func (b *FakeBroker) Calls() []BrokerCall {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]BrokerCall(nil), b.calls...)
}

// Consumer returns the consumer registered for queueName, or nil.
func (b *FakeBroker) Consumer(queueName string) amqp.MessageConsumer {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.consumers[queueName]
}
