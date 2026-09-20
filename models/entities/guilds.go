package entities

import amqp "github.com/kaellybot/kaelly-amqp"

// Guild is keyed by (id, game): the same Discord guild can host both bots, each
// with its own configuration row. Without the game, one bot registering a guild
// would make it look already known to the other.
type Guild struct {
	ID   string    `gorm:"primaryKey"`
	Game amqp.Game `gorm:"primaryKey"`
}
