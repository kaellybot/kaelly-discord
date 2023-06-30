package constants

const (
	AlignmentMinLevel = 0
	AlignmentMaxLevel = 100

	NeutralCityColor = 12506502
)

type AlignmentUserLevel struct {
	CityID   string
	OrderID  string
	Username string
	Level    int64
}
