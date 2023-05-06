package constants

type Graphist struct {
	Name string
	URL  string
}

func GetGraphist() Graphist {
	return Graphist{
		Name: "Elycann",
		URL:  "https://www.facebook.com/Elysdrawings",
	}
}
