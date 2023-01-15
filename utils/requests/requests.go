package requests

import (
	"context"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/utils/panics"
	"github.com/rs/zerolog/log"
)

func New(broker amqp.MessageBrokerInterface) *RequestManagerImpl {
	return &RequestManagerImpl{
		broker:   broker,
		requests: make(map[string]discordRequest),
	}
}

func GetBinding() amqp.Binding {
	return amqp.Binding{
		Exchange:   amqp.ExchangeAnswer,
		RoutingKey: AnswersRoutingKey,
		Queue:      AnswersQueueName,
	}
}

func (manager *RequestManagerImpl) Request(s *discordgo.Session, i *discordgo.InteractionCreate,
	routingKey string, message *amqp.RabbitMQMessage, callback RequestCallback,
	optionalProperties ...map[string]any) error {

	err := manager.broker.Publish(message, amqp.ExchangeRequest, routingKey, i.ID)
	if err != nil {
		return err
	}

	properties := make(map[string]any)
	if len(optionalProperties) > 0 {
		properties = optionalProperties[0]
	}

	manager.requests[i.ID] = discordRequest{
		session:     s,
		interaction: i,
		callback:    callback,
		properties:  properties,
	}
	return nil
}

func (manager *RequestManagerImpl) Listen() error {
	log.Info().Msgf("Listening request answers...")
	return manager.broker.Consume(AnswersQueueName, AnswersRoutingKey, manager.consume)
}

func (manager *RequestManagerImpl) consume(ctx context.Context, message *amqp.RabbitMQMessage, correlationId string) {
	request, found := manager.requests[correlationId]
	if found {
		defer panics.HandlePanic(request.session, request.interaction)
		delete(manager.requests, correlationId)
		request.callback(ctx, request.session, request.interaction, message, request.properties)
	}
}
