package constants

type CityType string

const (
	AlignmentMinLevel = 0
	AlignmentMaxLevel = 100

	CityTypeDark  CityType = "dark"
	CityTypeLight CityType = "light"

	NeutralCityColor = 12506502
)

type AlignmentUserLevel struct {
	CityID   string
	OrderID  string
	Username string
	Level    int64
}
