package constants

const (
	EmojiIDEffect      EmojiMiscID = "effect"
	EmojiIDRecipe      EmojiMiscID = "recipe"
	EmojiIDSet         EmojiMiscID = "set"
	EmojiIDNormalMap   EmojiMiscID = "normalMap"
	EmojiIDTacticalMap EmojiMiscID = "tacticalMap"

	EmojiTypeEquipment EmojiType = "Equipment"
	EmojiTypeBonusSet  EmojiType = "BonusSet"
	EmojiTypeMisc      EmojiType = "Miscellaneous"
)

type EmojiMiscID string
type EmojiType string
