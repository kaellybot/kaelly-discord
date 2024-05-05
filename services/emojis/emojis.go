package emojis

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
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

func (service *Impl) GetMiscEmoji(emojiMiscID constants.EmojiMiscID) *discordgo.ComponentEmoji {
	innerStore, found := service.emojiStore[constants.EmojiTypeMisc]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeEquipment)).
			Msgf("No miscellaneous type store found, returning empty emoji")
		return mapEmoji(entities.Emoji{})
	}

	emojiID := string(emojiMiscID)
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Msgf("No miscellaneous emoji found, returning empty emoji")
		return mapEmoji(entities.Emoji{})
	}

	return mapEmoji(emoji)
}

func (service *Impl) GetEquipmentEmoji(equipmentType amqp.EquipmentType) *discordgo.ComponentEmoji {
	innerStore, found := service.emojiStore[constants.EmojiTypeEquipment]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeEquipment)).
			Msgf("No equipment type store found, returning empty emoji")
		return mapEmoji(entities.Emoji{})
	}

	emojiID := equipmentType.String()
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Msgf("No equipment type emoji found, returning empty emoji")
		return mapEmoji(entities.Emoji{})
	}

	return mapEmoji(emoji)
}

func (service *Impl) GetSetBonusEmoji(equipedItemNumber, itemNumber int) *discordgo.ComponentEmoji {
	innerStore, found := service.emojiStore[constants.EmojiTypeBonusSet]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeBonusSet)).
			Msgf("No bonus set store found, returning empty emoji")
		return mapEmoji(entities.Emoji{})
	}

	emojiID := fmt.Sprintf("%v", len(innerStore)-(itemNumber-equipedItemNumber)+1)
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Str(constants.LogEmojiType, string(constants.EmojiTypeBonusSet)).
			Msgf("No bonus set emoji found, returning empty emoji")
		return mapEmoji(entities.Emoji{})
	}
	return mapEmoji(emoji)
}

func mapEmoji(emoji entities.Emoji) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{
		ID:   emoji.Snowflake,
		Name: emoji.Name,
	}
}
