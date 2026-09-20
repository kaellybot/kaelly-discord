package requests

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
)

const (
	AnswersExchange  = "amq.direct"
	AnswersQueueName = "answers"
)

type RequestManager interface {
	Request(s *discordgo.Session, i *discordgo.InteractionCreate, routingKey string,
		message *amqp.RabbitMQMessage, callback RequestCallback, properties ...map[string]any) error
	Listen()
	Reject(ctx amqp.Context, message *amqp.RabbitMQMessage, err error)
}

type RequestManagerImpl struct {
	broker   amqp.MessageBroker
	timeout  time.Duration
	mut      sync.Mutex
	requests map[string]discordRequest
}

type discordRequest struct {
	session     *discordgo.Session
	interaction *discordgo.InteractionCreate
	callback    RequestCallback
	properties  map[string]any
	// timer answers the caller if no reply ever comes back. Stopped as soon as
	// the request leaves the map, whether answered, refused or expired.
	timer *time.Timer
}

// RequestCallback handles the answer to a request. Returning an error means
// the caller could not be served: the manager turns it into a message for them.
type RequestCallback func(ctx context.Context, s *discordgo.Session,
	i *discordgo.InteractionCreate, message *amqp.RabbitMQMessage, properties map[string]any) error
