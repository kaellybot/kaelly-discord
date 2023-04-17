package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type FeedType struct {
	Id     string          `gorm:"primaryKey"`
	Labels []FeedTypeLabel `gorm:"foreignKey:FeedTypeId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type FeedTypeLabel struct {
	Locale     amqp.Language `gorm:"primaryKey"`
	FeedTypeId string        `gorm:"primaryKey"`
	Label      string
}

func (feedType FeedType) GetId() string {
	return feedType.Id
}

func (feedType FeedType) GetLabels() map[amqp.Language]string {
	labels := make(map[amqp.Language]string)

	for _, label := range feedType.Labels {
		labels[label.Locale] = label.Label
	}

	return labels
}
