package application

import (
	"errors"
	"sync"
	"testing"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/services/discord"
	"github.com/kaellybot/kaelly-discord/utils/databases"
	"github.com/kaellybot/kaelly-discord/utils/requests"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// journal records lifecycle calls across all fakes, in order.
type journal struct {
	mu    sync.Mutex
	steps []string
}

func (j *journal) add(step string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.steps = append(j.steps, step)
}

type fakeBroker struct {
	amqp.MessageBroker
	j      *journal
	runErr error
}

func (f *fakeBroker) Run() error { f.j.add("broker.Run"); return f.runErr }
func (f *fakeBroker) Shutdown()  { f.j.add("broker.Shutdown") }

type fakeDB struct {
	databases.MySQLConnection
	j *journal
}

func (f *fakeDB) Shutdown() { f.j.add("db.Shutdown") }

type fakeServer struct {
	name string
	j    *journal
}

func (f *fakeServer) ListenAndServe() { f.j.add(f.name + ".ListenAndServe") }
func (f *fakeServer) Shutdown()       { f.j.add(f.name + ".Shutdown") }

type fakeDiscord struct {
	discord.Service
	j         *journal
	listenErr error
}

func (f *fakeDiscord) Listen() error { f.j.add("discord.Listen"); return f.listenErr }
func (f *fakeDiscord) Shutdown()     { f.j.add("discord.Shutdown") }

type fakeRequests struct {
	requests.RequestManager
	j *journal
}

func (f *fakeRequests) Listen() { f.j.add("requests.Listen") }

var _ requests.RequestManager = (*fakeRequests)(nil)

func newApp(brokerErr, discordErr error) (*Impl, *journal) {
	j := &journal{}
	return &Impl{
		broker:         &fakeBroker{j: j, runErr: brokerErr},
		db:             &fakeDB{j: j},
		probes:         &fakeServer{name: "probes", j: j},
		prom:           &fakeServer{name: "prom", j: j},
		discordService: &fakeDiscord{j: j, listenErr: discordErr},
		requestManager: &fakeRequests{j: j},
	}, j
}

func TestRun_StartsEverythingInDependencyOrder(t *testing.T) {
	app, j := newApp(nil, nil)

	require.NoError(t, app.Run())

	assert.Equal(t, []string{
		"probes.ListenAndServe",
		"prom.ListenAndServe",
		"broker.Run",
		"requests.Listen",
		"discord.Listen",
	}, j.steps, "reply consumers must be listening before Discord interactions can produce requests")
}

func TestRun_BrokerFailureStopsBeforeListening(t *testing.T) {
	brokerErr := errors.New("rabbitmq unreachable")
	app, j := newApp(brokerErr, nil)

	err := app.Run()

	require.ErrorIs(t, err, brokerErr)
	assert.Equal(t, []string{
		"probes.ListenAndServe",
		"prom.ListenAndServe",
		"broker.Run",
	}, j.steps, "neither the request manager nor Discord are started")
}

func TestRun_DiscordFailureIsReturned(t *testing.T) {
	discordErr := errors.New("invalid token")
	app, j := newApp(nil, discordErr)

	err := app.Run()

	require.ErrorIs(t, err, discordErr)
	assert.Equal(t, "discord.Listen", j.steps[len(j.steps)-1])
}

func TestShutdown_StopsIncomingTrafficFirstAndProbesLast(t *testing.T) {
	app, j := newApp(nil, nil)

	app.Shutdown()

	assert.Equal(t, []string{
		"discord.Shutdown",
		"broker.Shutdown",
		"db.Shutdown",
		"prom.Shutdown",
		"probes.Shutdown",
	}, j.steps)
}

func TestNew_UnreachableDatabaseFailsFast(t *testing.T) {
	// Port 1 on loopback refuses connections immediately; amqp.New does not dial.
	viper.Set(constants.MySQLURL, "127.0.0.1:1")
	viper.Set(constants.RabbitMQAddress, "amqp://127.0.0.1:1")
	t.Cleanup(func() {
		viper.Set(constants.MySQLURL, nil)
		viper.Set(constants.RabbitMQAddress, nil)
	})

	app, err := New()

	require.Error(t, err)
	assert.Nil(t, app)
}
