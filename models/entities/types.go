package entities

import (
	amqp "github.com/kaellybot/kaelly-amqp"
)

type LabelledEntity interface {
	GetID() string
	GetLabels() map[amqp.Language]string
}
