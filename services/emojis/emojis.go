package emojis

import (
	"fmt"
	"strings"

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

func (service *Impl) GetMiscStringEmoji(emojiMiscID constants.EmojiMiscID) string {
	innerStore, found := service.emojiStore[constants.EmojiTypeMisc]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeMisc)).
			Msgf("No miscellaneous type store found, returning empty emoji")
		return mapEmojiString(nil)
	}

	emojiID := string(emojiMiscID)
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Msgf("No miscellaneous emoji found, returning empty emoji")
		return mapEmojiString(nil)
	}

	return mapEmojiString(&emoji)
}

func (service *Impl) GetMiscEmoji(emojiMiscID constants.EmojiMiscID) *discordgo.ComponentEmoji {
	innerStore, found := service.emojiStore[constants.EmojiTypeMisc]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeMisc)).
			Msgf("No miscellaneous type store found, returning empty emoji")
		return mapEmoji(nil)
	}

	emojiID := string(emojiMiscID)
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Msgf("No miscellaneous emoji found, returning empty emoji")
		return mapEmoji(nil)
	}

	return mapEmoji(&emoji)
}

func (service *Impl) GetEntityStringEmoji(entityID string, entityType constants.EmojiEntity) string {
	innerStore, found := service.emojiStore[constants.EmojiType(entityType)]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(entityType)).
			Msgf("No entity emoji type store found, returning empty emoji")
		return mapEmojiString(nil)
	}

	emoji, found := innerStore[entityID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, entityID).
			Msgf("No entity emoji found, returning empty emoji")
		return mapEmojiString(nil)
	}

	return mapEmojiString(&emoji)
}

func (service *Impl) GetEntityEmoji(entityID string, entityType constants.EmojiEntity) *discordgo.ComponentEmoji {
	innerStore, found := service.emojiStore[constants.EmojiType(entityType)]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(entityType)).
			Msgf("No entity emoji type store found, returning empty emoji")
		return mapEmoji(nil)
	}

	emoji, found := innerStore[entityID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, entityID).
			Msgf("No entity emoji found, returning empty emoji")
		return mapEmoji(nil)
	}

	return mapEmoji(&emoji)
}

func (service *Impl) GetEquipmentStringEmoji(equipmentType amqp.EquipmentType) string {
	innerStore, found := service.emojiStore[constants.EmojiTypeEquipment]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeEquipment)).
			Msgf("No equipment type store found, returning empty emoji")
		return mapEmojiString(nil)
	}

	emojiID := equipmentType.String()
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Msgf("No equipment type emoji found, returning empty emoji")
		return mapEmojiString(nil)
	}

	return mapEmojiString(&emoji)
}

func (service *Impl) GetEquipmentEmoji(equipmentType amqp.EquipmentType) *discordgo.ComponentEmoji {
	innerStore, found := service.emojiStore[constants.EmojiTypeEquipment]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeEquipment)).
			Msgf("No equipment type store found, returning empty emoji")
		return mapEmoji(nil)
	}

	emojiID := equipmentType.String()
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Msgf("No equipment type emoji found, returning empty emoji")
		return mapEmoji(nil)
	}

	return mapEmoji(&emoji)
}

func (service *Impl) GetItemTypeStringEmoji(itemType amqp.ItemType) string {
	innerStore, found := service.emojiStore[constants.EmojiTypeItem]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeItem)).
			Msgf("No item type store found, returning empty emoji")
		return mapEmojiString(nil)
	}

	emojiID := itemType.String()
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Msgf("No item type emoji found, returning empty emoji")
		return mapEmojiString(nil)
	}

	return mapEmojiString(&emoji)
}

func (service *Impl) GetItemTypeEmoji(itemType amqp.ItemType) *discordgo.ComponentEmoji {
	innerStore, found := service.emojiStore[constants.EmojiTypeItem]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeItem)).
			Msgf("No item type store found, returning empty emoji")
		return mapEmoji(nil)
	}

	emojiID := itemType.String()
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Msgf("No item type emoji found, returning empty emoji")
		return mapEmoji(nil)
	}

	return mapEmoji(&emoji)
}

func (service *Impl) GetSetBonusEmoji(equipedItemNumber int) *discordgo.ComponentEmoji {
	innerStore, found := service.emojiStore[constants.EmojiTypeBonusSet]
	if !found {
		log.Warn().
			Str(constants.LogEmojiType, string(constants.EmojiTypeBonusSet)).
			Msgf("No bonus set store found, returning empty emoji")
		return mapEmoji(nil)
	}

	emojiID := fmt.Sprintf("%v", equipedItemNumber)
	emoji, found := innerStore[emojiID]
	if !found {
		log.Warn().
			Str(constants.LogEntity, emojiID).
			Str(constants.LogEmojiType, string(constants.EmojiTypeBonusSet)).
			Msgf("No bonus set emoji found, returning empty emoji")
		return mapEmoji(nil)
	}
	return mapEmoji(&emoji)
}

func mapEmoji(emoji *entities.Emoji) *discordgo.ComponentEmoji {
	if emoji == nil {
		return nil
	}

	return &discordgo.ComponentEmoji{
		ID:   emoji.Snowflake,
		Name: emoji.Name,
	}
}

func mapEmojiString(emoji *entities.Emoji) string {
	if emoji != nil {
		if len(strings.TrimSpace(emoji.ID)) > 0 {
			return fmt.Sprintf("<:%v:%v>", emoji.DiscordName, emoji.Snowflake)
		}

		return emoji.Name
	}

	return ""
}
