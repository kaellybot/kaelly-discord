package requests

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/commands"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/utils/discord"
	"github.com/kaellybot/kaelly-discord/utils/panics"
	"github.com/rs/zerolog/log"
)

func New(broker amqp.MessageBroker) *RequestManagerImpl {
	return &RequestManagerImpl{
		broker:   broker,
		timeout:  constants.GetRequestTimeout(),
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
		// Without this the entry would sit in the map forever and the caller
		// would stay on "thinking..." until Discord's token expires.
		timer: time.AfterFunc(manager.timeout, func() { manager.expire(i.ID) }),
	}
	manager.mut.Unlock()

	err := manager.broker.Request(message, amqp.ExchangeRequest, routingKey, i.ID, AnswersQueueName)
	if err != nil {
		manager.take(i.ID)
		// Classified here rather than at each caller: a broker refusing a
		// publish is the textbook "try again later".
		return fmt.Errorf("%w: %w", commands.ErrUnavailable, err)
	}

	return nil
}

func (manager *RequestManagerImpl) Listen() {
	log.Info().Msgf("Listening request answers...")
	manager.broker.Consume(AnswersQueueName, manager.consume)
}

func (manager *RequestManagerImpl) consume(ctx amqp.Context, message *amqp.RabbitMQMessage) {
	request, found := manager.take(ctx.CorrelationID)
	if !found {
		// Already answered, refused or expired: a late reply is ignored.
		return
	}

	defer panics.HandlePanic(request.session, request.interaction)
	if err := request.callback(ctx, request.session, request.interaction,
		message, request.properties); err != nil {
		discord.AnswerError(request.session, request.interaction, err)
	}
}

// Reject answers the caller of a reply the broker refused, typically one whose
// game is not set. Without it the caller would wait out the whole timeout for an
// answer that has already been thrown away.
func (manager *RequestManagerImpl) Reject(ctx amqp.Context, message *amqp.RabbitMQMessage,
	err error) {
	request, found := manager.take(ctx.CorrelationID)
	if !found {
		return
	}

	log.Error().Err(err).
		Str(constants.LogCorrelationID, ctx.CorrelationID).
		Str(constants.LogGame, message.GetGame().String()).
		Msgf("Answer refused by the broker, telling the caller")
	discord.AnswerError(request.session, request.interaction,
		fmt.Errorf("%w: %w", commands.ErrUnavailable, err))
}

func (manager *RequestManagerImpl) expire(correlationID string) {
	request, found := manager.take(correlationID)
	if !found {
		return
	}

	log.Warn().
		Str(constants.LogCorrelationID, correlationID).
		Msgf("No answer received in %v, telling the caller", manager.timeout)
	discord.AnswerError(request.session, request.interaction,
		fmt.Errorf("%w: no answer received in %v", commands.ErrUnavailable, manager.timeout))
}

// take removes a request from the map and stops its timer, so that exactly one of
// consume, Reject and expire ever answers the caller.
//
// Discord handlers register requests from their own goroutines while answers are
// consumed from another: the map is only touched under lock, and the callback runs
// outside of it.
func (manager *RequestManagerImpl) take(correlationID string) (discordRequest, bool) {
	manager.mut.Lock()
	request, found := manager.requests[correlationID]
	delete(manager.requests, correlationID)
	manager.mut.Unlock()

	if found && request.timer != nil {
		request.timer.Stop()
	}

	return request, found
}
