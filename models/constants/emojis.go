package constants

const (
	EmojiIDAlmanax     EmojiMiscID = "almanax"
	EmojiIDCalendar    EmojiMiscID = "calendar"
	EmojiIDCost        EmojiMiscID = "cost"
	EmojiIDCritical    EmojiMiscID = "critical"
	EmojiIDEffect      EmojiMiscID = "effect"
	EmojiIDFirst       EmojiMiscID = "first"
	EmojiIDGame        EmojiMiscID = "dofusGame"
	EmojiIDGithub      EmojiMiscID = "github"
	EmojiIDGlobal      EmojiMiscID = "global"
	EmojiIDKama        EmojiMiscID = "kama"
	EmojiIDLast        EmojiMiscID = "last"
	EmojiIDNext        EmojiMiscID = "next"
	EmojiIDNormalMap   EmojiMiscID = "normalMap"
	EmojiIDPrevious    EmojiMiscID = "previous"
	EmojiIDRange       EmojiMiscID = "range"
	EmojiIDRecipe      EmojiMiscID = "recipe"
	EmojiIDRSS         EmojiMiscID = "rss"
	EmojiIDScroll      EmojiMiscID = "scroll"
	EmojiIDTacticalMap EmojiMiscID = "tacticalMap"
	EmojiIDTwitter     EmojiMiscID = "twitter"

	EmojiTypeBonusSet  EmojiType = "BonusSet"
	EmojiTypeEquipment EmojiType = "Equipment"
	EmojiTypeItem      EmojiType = "Item"
	EmojiTypeMisc      EmojiType = "Miscellaneous"

	EmojiTypeCharacteristic EmojiEntity = "Characteristic"
	EmojiTypeCity           EmojiEntity = "City"
	EmojiTypeJob            EmojiEntity = "Job"
	EmojiTypeOrder          EmojiEntity = "Order"
	EmojiTypeServer         EmojiEntity = "Server"
	EmojiTypeTransportType  EmojiEntity = "TransportType"
)

type EmojiMiscID string
type EmojiType string
type EmojiEntity EmojiType
