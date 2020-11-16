package neatgo

// Connection ...
type Connection struct {
	In         int
	Out        int
	Weight     float64
	Enabled    bool
	Innovation int64
}

// Clone ...
func (o Connection) Clone() *Connection {
	return &Connection{
		In:         o.In,
		Out:        o.Out,
		Weight:     o.Weight,
		Enabled:    o.Enabled,
		Innovation: o.Innovation,
	}
}
