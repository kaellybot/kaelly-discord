package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type FeedType struct {
	ID      string          `gorm:"primaryKey"`
	Labels  []FeedTypeLabel `gorm:"foreignKey:FeedTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Sources []FeedSource    `gorm:"foreignKey:FeedTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type FeedTypeLabel struct {
	Locale     amqp.Language `gorm:"primaryKey"`
	FeedTypeID string        `gorm:"primaryKey"`
	Label      string
}

type FeedSource struct {
	Locale        amqp.Language `gorm:"primaryKey"`
	Game          amqp.Game     `gorm:"primaryKey"`
	FeedTypeID    string        `gorm:"primaryKey"`
	NewsChannelID string
}

func (feedType FeedType) GetID() string {
	return feedType.ID
}

func (feedType FeedType) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range feedType.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
