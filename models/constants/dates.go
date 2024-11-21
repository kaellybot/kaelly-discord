package constants

const (
	DiscordDateOnlyFormat = "02/01/2006"
)

func SupportedDateFormats() []string {
	return []string{
		"02-01-2006",
		"02/01/2006",
		"02 01 2006",
		"2006-01-02",
		"2006/01/02",
		"2006 01 02",
	}
}
