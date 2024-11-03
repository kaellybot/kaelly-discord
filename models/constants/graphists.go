package constants

type Graphist struct {
	Name string
	URL  string
}

func GetGraphistElycann() Graphist {
	return Graphist{
		Name: "Elycann",
		URL:  "https://www.facebook.com/Elysdrawings",
	}
}

func GetGraphistColibry() Graphist {
	return Graphist{
		Name: "Colibry",
		URL:  "https://x.com/AmelBencivenni",
	}
}
