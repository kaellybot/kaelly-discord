package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
)

type LabelledEntity interface {
	GetId() string
	GetLabels() map[amqp.Language]string
}
