package emojis

import (
	"fmt"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-discord/models/constants"
	"github.com/kaellybot/kaelly-discord/models/entities"
	repository "github.com/kaellybot/kaelly-discord/repositories/emojis"
	"github.com/rs/zerolog/log"
)

func New(repository repository.Repository) (*Impl, error) {
	emojis, err := repository.GetEmojis()
	if err != nil {
		return nil, err
	}

	log.Info().
		Int(constants.LogEntityCount, len(emojis)).
		Msgf("Emojis loaded")

	emojiStore := make(map[constants.EmojiType]emojiStore)
	for _, emoji := range emojis {
		innerStore, found := emojiStore[emoji.Type]
		if !found {
			innerStore = make(map[string]entities.Emoji)
			emojiStore[emoji.Type] = innerStore
		}

		innerStore[emoji.ID] = emoji
	}

	return &Impl{
		emojiStore: emojiStore,
		repository: repository,
	}, nil
}

func (service *Impl) GetEquipmentEmoji(equipmentType amqp.EquipmentType) entities.Emoji {
	innerStore, found := service.emojiStore[constants.EmojiTypeEquipment]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeEquipment)).
			Msgf("No equipment type store found, returning empty emoji")
		return entities.Emoji{Type: constants.EmojiTypeEquipment}
	}

	emojiID := equipmentType.String()
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Msgf("No equipment type found, returning empty emoji")
		return entities.Emoji{Type: constants.EmojiTypeEquipment}
	}

	return emoji
}

func (service *Impl) GetSetBonusEmoji(equipedItemNumber, itemNumber int) entities.Emoji {
	innerStore, found := service.emojiStore[constants.EmojiTypeBonusSet]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeBonusSet)).
			Msgf("No bonus set store found, returning empty emoji")
		return entities.Emoji{Type: constants.EmojiTypeBonusSet}
	}

	emojiID := fmt.Sprintf("%v", len(innerStore)-(itemNumber-equipedItemNumber)+1)
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Str(constants.LogEmojiType, string(constants.EmojiTypeBonusSet)).
			Msgf("No bonus set emoji found, returning empty emoji")
		return entities.Emoji{Type: constants.EmojiTypeBonusSet}
	}
	return emoji
}
