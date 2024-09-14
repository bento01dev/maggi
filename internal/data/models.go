package data

type Profile struct {
	ID   int
	Name string
}

type DetailType string

func (d DetailType) String() string {
	return string(d)
}

const (
	AliasDetail DetailType = "alias"
	EnvDetail   DetailType = "env"
)

type Detail struct {
	ID         int
	Key        string
	Value      string
	DetailType DetailType
	ProfileID  int
}
