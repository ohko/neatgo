package neatgo

// ...
const (
	NodeTypeInput  = "input"
	NodeTypeHidden = "hidden"
	NodeTypeOutput = "output"
)

// Node ...
type Node struct {
	Index    int
	Type     string
	Activate string
	Value    float64 `json:"-"`
}
