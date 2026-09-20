package requests

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutils.LoadI18n()
	m.Run()
}

type callback struct {
	calls      int
	ctx        context.Context
	session    *discordgo.Session
	event      *discordgo.InteractionCreate
	message    *amqp.RabbitMQMessage
	properties map[string]any
}

func (c *callback) handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate,
	message *amqp.RabbitMQMessage, properties map[string]any) {
	c.calls++
	c.ctx, c.session, c.event, c.message, c.properties = ctx, s, i, message, properties
}

func interaction(id string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: id, AppID: "app", Token: "tok", Type: discordgo.InteractionApplicationCommand,
		Locale: discordgo.French, Data: discordgo.ApplicationCommandInteractionData{Name: "almanax"},
	}}
}

// listening returns a manager plugged on a fake broker, and the consumer it registered.
func listening(t *testing.T, broker *testutils.FakeBroker) (*RequestManagerImpl, amqp.MessageConsumer) {
	t.Helper()
	manager := New(broker)
	manager.Listen()
	consumer := broker.Consumer(AnswersQueueName)
	require.NotNil(t, consumer, "Listen must consume the answers queue")
	return manager, consumer
}

func answer(correlationID string) amqp.Context {
	return amqp.Context{Context: context.Background(), CorrelationID: correlationID}
}

func TestGetBinding_AnswersQueueOnDirectExchange(t *testing.T) {
	assert.Equal(t, amqp.Binding{Exchange: "amq.direct", RoutingKey: "", Queue: "answers"}, GetBinding())
}

func TestRequest_PublishesWithInteractionIDAsCorrelation(t *testing.T) {
	broker := &testutils.FakeBroker{}
	manager, _ := listening(t, broker)
	message := &amqp.RabbitMQMessage{Language: amqp.Language_FR}

	err := manager.Request(nil, interaction("i1"), "requests.encyclopedias", message, (&callback{}).handle)

	require.NoError(t, err)
	assert.Equal(t, []testutils.BrokerCall{{
		Method: "Request", Message: message, Exchange: amqp.ExchangeRequest,
		RoutingKey: "requests.encyclopedias", CorrelationID: "i1", ReplyTo: AnswersQueueName,
	}}, broker.Calls())
}

func TestRequest_AnswerRoutedToCallbackWithContext(t *testing.T) {
	broker := &testutils.FakeBroker{}
	manager, consume := listening(t, broker)
	session, fake := testutils.NewDiscordSession(t)
	event := interaction("i1")
	cb := &callback{}
	properties := map[string]any{"members": 3}
	require.NoError(t, manager.Request(session, event, "rk", &amqp.RabbitMQMessage{}, cb.handle, properties))

	reply := &amqp.RabbitMQMessage{Language: amqp.Language_DE}
	ctx := answer("i1")
	consume(ctx, reply)

	assert.Equal(t, 1, cb.calls)
	assert.Equal(t, ctx, cb.ctx, "the AMQP context is handed over")
	assert.Same(t, session, cb.session)
	assert.Same(t, event, cb.event)
	assert.Same(t, reply, cb.message)
	assert.Equal(t, properties, cb.properties)
	assert.Empty(t, fake.Requests(), "the manager itself never talks to Discord on success")
}

func TestRequest_OnlyFirstOptionalPropertiesAreKept(t *testing.T) {
	broker := &testutils.FakeBroker{}
	manager, consume := listening(t, broker)
	cb := &callback{}
	require.NoError(t, manager.Request(nil, interaction("i1"), "rk", &amqp.RabbitMQMessage{}, cb.handle,
		map[string]any{"a": 1}, map[string]any{"b": 2}))

	consume(answer("i1"), &amqp.RabbitMQMessage{})

	assert.Equal(t, map[string]any{"a": 1}, cb.properties)
}

func TestRequest_WithoutPropertiesCallbackGetsEmptyWritableMap(t *testing.T) {
	broker := &testutils.FakeBroker{}
	manager, consume := listening(t, broker)
	cb := &callback{}
	require.NoError(t, manager.Request(nil, interaction("i1"), "rk", &amqp.RabbitMQMessage{}, cb.handle))

	consume(answer("i1"), &amqp.RabbitMQMessage{})

	require.NotNil(t, cb.properties)
	assert.Empty(t, cb.properties)
	assert.NotPanics(t, func() { cb.properties["k"] = "v" })
}

func TestRequest_BrokerFailureIsReturnedAndNothingIsAwaited(t *testing.T) {
	errBroker := errors.New("channel closed")
	broker := &testutils.FakeBroker{Err: errBroker}
	manager, consume := listening(t, broker)
	cb := &callback{}

	err := manager.Request(nil, interaction("i1"), "rk", &amqp.RabbitMQMessage{}, cb.handle)

	require.ErrorIs(t, err, errBroker)
	consume(answer("i1"), &amqp.RabbitMQMessage{})
	assert.Zero(t, cb.calls, "a request that was never sent cannot be answered")
}

func TestConsume_UnknownCorrelationIsIgnored(t *testing.T) {
	broker := &testutils.FakeBroker{}
	manager, consume := listening(t, broker)
	cb := &callback{}
	require.NoError(t, manager.Request(nil, interaction("i1"), "rk", &amqp.RabbitMQMessage{}, cb.handle))

	consume(answer("other"), &amqp.RabbitMQMessage{})
	consume(answer(""), &amqp.RabbitMQMessage{})

	assert.Zero(t, cb.calls)
	consume(answer("i1"), &amqp.RabbitMQMessage{})
	assert.Equal(t, 1, cb.calls, "the pending request survived unrelated answers")
}

func TestConsume_AnswerIsDeliveredOnlyOnce(t *testing.T) {
	broker := &testutils.FakeBroker{}
	manager, consume := listening(t, broker)
	cb := &callback{}
	require.NoError(t, manager.Request(nil, interaction("i1"), "rk", &amqp.RabbitMQMessage{}, cb.handle))

	consume(answer("i1"), &amqp.RabbitMQMessage{})
	consume(answer("i1"), &amqp.RabbitMQMessage{})

	assert.Equal(t, 1, cb.calls, "duplicate deliveries are dropped")
}

func TestConsume_ConcurrentRequestsAreRoutedIndependently(t *testing.T) {
	broker := &testutils.FakeBroker{}
	manager, consume := listening(t, broker)
	first, second := &callback{}, &callback{}
	require.NoError(t, manager.Request(nil, interaction("i1"), "rk", &amqp.RabbitMQMessage{}, first.handle))
	require.NoError(t, manager.Request(nil, interaction("i2"), "rk", &amqp.RabbitMQMessage{}, second.handle))

	consume(answer("i2"), &amqp.RabbitMQMessage{})

	assert.Zero(t, first.calls)
	assert.Equal(t, 1, second.calls)
	assert.Equal(t, "i2", second.event.ID)
}

func TestConsume_PanickingCallbackIsRecoveredAndUserIsNotified(t *testing.T) {
	broker := &testutils.FakeBroker{}
	manager, consume := listening(t, broker)
	session, fake := testutils.NewDiscordSession(t)
	calls := 0
	panicking := func(context.Context, *discordgo.Session, *discordgo.InteractionCreate,
		*amqp.RabbitMQMessage, map[string]any) {
		calls++
		panic("nil answer")
	}
	require.NoError(t, manager.Request(session, interaction("i1"), "rk", &amqp.RabbitMQMessage{}, panicking))

	require.NotPanics(t, func() { consume(answer("i1"), &amqp.RabbitMQMessage{}) })

	req := fake.Last(t)
	assert.Equal(t, "PATCH", req.Method)
	assert.Contains(t, req.Path, "/webhooks/app/tok/messages/@original")
	payload := fake.LastEdit(t)
	require.NotNil(t, payload.Content)
	assert.Contains(t, testutils.Translations(t, discordgo.French, "panic"), *payload.Content)

	consume(answer("i1"), &amqp.RabbitMQMessage{})
	assert.Equal(t, 1, calls, "the request is forgotten even when its callback crashed")
}

func TestRequest_ConcurrentWithConsumeIsRaceFree(t *testing.T) {
	// Requests are registered from the goroutine discordgo spawns per event
	// while answers are consumed from the AMQP one: without a lock this is a
	// data race, and a lost race crashes the process on "concurrent map writes".
	broker := &testutils.FakeBroker{}
	manager, consume := listening(t, broker)
	callbacks := make([]*callback, 64)

	var waitGroup sync.WaitGroup
	for index := range callbacks {
		callbacks[index] = &callback{}
		id := strconv.Itoa(index)
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			assert.NoError(t, manager.Request(nil, interaction(id), "rk",
				&amqp.RabbitMQMessage{}, callbacks[index].handle))
		}()
		go func() {
			defer waitGroup.Done()
			consume(answer(id), &amqp.RabbitMQMessage{})
		}()
	}
	waitGroup.Wait()

	// Answers may land before their request was registered, so replay them all:
	// every request is still pending or already delivered, never lost.
	for index, cb := range callbacks {
		consume(answer(strconv.Itoa(index)), &amqp.RabbitMQMessage{})
		assert.Equal(t, 1, cb.calls, "request %d answered exactly once", index)
	}
}

func TestRequest_AnswerArrivingBeforeRequestReturnsIsDelivered(t *testing.T) {
	// The broker replies from inside Request: the pending entry must already be
	// registered by then, otherwise the answer is dropped and the user is stuck
	// on "thinking...".
	broker := &testutils.FakeBroker{}
	manager, consume := listening(t, broker)
	cb := &callback{}
	broker.OnRequest = func(correlationID string) {
		consume(answer(correlationID), &amqp.RabbitMQMessage{})
	}

	require.NoError(t, manager.Request(nil, interaction("i1"), "rk", &amqp.RabbitMQMessage{}, cb.handle))

	assert.Equal(t, 1, cb.calls, "the answer was routed even though it raced Request")
}
