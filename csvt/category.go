package csvt

type category string

const (
	MAP category = "MAP"
	ARR category = "ARR"
	STR category = "STR"
	OBJ category = "OBJ"
)

func (m category) String() string {
	return string(m)
}
