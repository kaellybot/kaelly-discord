package requests

import (
	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/utils/panics"
	"github.com/rs/zerolog/log"
)

func New(broker amqp.MessageBroker) *RequestManagerImpl {
	return &RequestManagerImpl{
		broker:   broker,
		requests: make(map[string]discordRequest),
	}
}

func GetBinding() amqp.Binding {
	return amqp.Binding{
		Exchange:   AnswersExchange,
		RoutingKey: "",
		Queue:      AnswersQueueName,
	}
}

func (manager *RequestManagerImpl) Request(s *discordgo.Session, i *discordgo.InteractionCreate,
	routingKey string, message *amqp.RabbitMQMessage, callback RequestCallback,
	optionalProperties ...map[string]any) error {
	properties := make(map[string]any)
	if len(optionalProperties) > 0 {
		properties = optionalProperties[0]
	}

	// Registered before publishing: an answer can come back before Request
	// returns, and consume would then find nothing to call.
	manager.mut.Lock()
	manager.requests[i.ID] = discordRequest{
		session:     s,
		interaction: i,
		callback:    callback,
		properties:  properties,
	}
	manager.mut.Unlock()

	err := manager.broker.Request(message, amqp.ExchangeRequest, routingKey, i.ID, AnswersQueueName)
	if err != nil {
		manager.mut.Lock()
		delete(manager.requests, i.ID)
		manager.mut.Unlock()
		return err
	}

	return nil
}

func (manager *RequestManagerImpl) Listen() {
	log.Info().Msgf("Listening request answers...")
	manager.broker.Consume(AnswersQueueName, manager.consume)
}

func (manager *RequestManagerImpl) consume(ctx amqp.Context, message *amqp.RabbitMQMessage) {
	// Discord handlers register requests from their own goroutines while
	// answers are consumed from this one: the map is only touched under lock,
	// and the callback runs outside of it.
	manager.mut.Lock()
	request, found := manager.requests[ctx.CorrelationID]
	delete(manager.requests, ctx.CorrelationID)
	manager.mut.Unlock()

	if found {
		defer panics.HandlePanic(request.session, request.interaction)
		request.callback(ctx, request.session, request.interaction, message, request.properties)
	}
}
