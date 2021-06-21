package neatgo

import (
	"encoding/json"
	"math"
)

// Genome ...
type Genome struct {
	Population  *Population `json:"-"`
	Nodes       map[int]*Node
	Connections []*Connection
	NextNodeID  int
	Fitness     float64
}

// NewGenome ...
func NewGenome(population *Population) (*Genome, error) {
	return &Genome{
		Population: population,
		Nodes:      make(map[int]*Node),
	}, nil
}

func (o *Genome) init() {
	for i := 0; i < o.Population.inputNumber; i++ {
		o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeInput, Value: NeatRandom(-1, 1)}
		o.NextNodeID++
	}
	for j := 0; j < o.Population.outputNumber; j++ {
		o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeOutput, Value: 0, Activate: "LOGISTIC"}

		if o.Population.Options.AllConnection {
			for i := 0; i < o.Population.inputNumber; i++ {
				o.Connections = append(o.Connections, &Connection{
					In:         o.Nodes[i].Index,
					Out:        o.NextNodeID,
					Weight:     NeatRandom(-1, 1),
					Enabled:    true,
					Innovation: o.Population.nextInnovationID,
				})
				o.Population.nextInnovationID++
			}
		} else {
			o.Connections = append(o.Connections, &Connection{
				In:         RandIntn(0, o.Population.inputNumber-1),
				Out:        o.NextNodeID,
				Weight:     NeatRandom(-1, 1),
				Enabled:    true,
				Innovation: o.Population.nextInnovationID,
			})
			o.Population.nextInnovationID++
		}

		o.NextNodeID++
	}
	for i := 0; i < o.Population.hiddenNumber; i++ {
		o.addNode()
	}
}
func (o *Genome) nextGeneration(n, dis int) {
	o.mutateWeight(n, dis)
	if dis < o.Population.Options.MaxDistance {
		return
	}

	div := math.Max(1, o.Population.Options.AddNode+o.Population.Options.AddConnection)
	r, r1, r2 := NeatRandom(0, 1), o.Population.Options.AddNode, o.Population.Options.AddConnection
	if r < r1/div {
		o.addNode()
	} else if r < (r1+r2)/div {
		o.addConnection()
	}
}
func (o *Genome) mutateWeight(n, dis int) {
	r := 0.0
	for i := 0; i < len(o.Connections); i++ {
		r = NeatRandom(0, 1)
		if r < 0.01 {
			o.Connections[i].Weight = NeatRandom(-1, 1)
		} else if r < o.Population.Options.MutateWeight {
			o.Connections[i].Weight += NeatRandom(-1, 1) * math.Min(float64(n+1), 10)
		}
	}
}
func (o *Genome) crossover(b *Genome) *Genome {
	for m := range o.Connections {
		for n := range b.Connections {
			if b.Connections[n].Innovation != o.Connections[m].Innovation {
				continue
			}

			o.Connections[m].Enabled = true
			b.Connections[n].Enabled = true

			if !o.Connections[m].Enabled || !b.Connections[n].Enabled {
				if NeatRandom(0, 1) < 0.75 {
					o.Connections[m].Enabled = false
					b.Connections[n].Enabled = false
				}
			}

			if NeatRandom(0, 1) < 0.5 {
				x := b.Connections[n].Weight
				b.Connections[n].Weight = o.Connections[m].Weight
				o.Connections[m].Weight = x
			}
		}
	}

	return o
}
func (o *Genome) addConnection() {
	for in := range o.Nodes {
		if o.Nodes[in].Type == NodeTypeOutput {
			continue
		}
		for out := range o.Nodes {
			if o.Nodes[out].Type == NodeTypeInput || out <= in {
				continue
			}

			found := false
			for _, c := range o.Connections {
				if c.In == in && c.Out == out {
					found = true
					break
				}
			}
			if found {
				continue
			}

			o.Connections = append(o.Connections, &Connection{
				In:         in,
				Out:        out,
				Weight:     NeatRandom(-1, 1),
				Enabled:    true,
				Innovation: o.Population.nextInnovationID,
			})
			o.Population.nextInnovationID++
			return
		}
	}
}
func (o *Genome) addNode() {
	if len(o.Nodes) > o.Population.Options.MaxNode {
		return
	}
	outs := []*Connection{}
	for a := range o.Connections {
		if o.Nodes[o.Connections[a].Out].Type == NodeTypeOutput {
			outs = append(outs, o.Connections[a])
		}
	}

	o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeHidden, Value: 0, Activate: randActivateFunc()}

	c := RandIntn(0, len(outs)-1)
	outs[c].Enabled = false
	o.Connections = append(o.Connections, &Connection{
		In:         outs[c].In,
		Out:        o.NextNodeID,
		Weight:     NeatRandom(-1, 1),
		Enabled:    true,
		Innovation: o.Population.nextInnovationID,
	})
	o.Population.nextInnovationID++
	o.Connections = append(o.Connections, &Connection{
		In:         o.NextNodeID,
		Out:        outs[c].Out,
		Weight:     NeatRandom(-1, 1),
		Enabled:    true,
		Innovation: o.Population.nextInnovationID,
	})
	o.Population.nextInnovationID++

	o.NextNodeID++
}

// ToJSON ...
func (o *Genome) ToJSON() string {
	bs, _ := json.Marshal(o)
	return string(bs)
}

// LoadJSON ...
func (o *Genome) LoadJSON(js string) error {
	return json.Unmarshal([]byte(js), o)
}
func (o *Genome) clone() *Genome {
	n, _ := NewGenome(o.Population)
	n.NextNodeID = o.NextNodeID
	n.Fitness = o.Fitness
	for k, v := range o.Nodes {
		n.Nodes[k] = &Node{Index: v.Index, Type: v.Type, Value: v.Value, Activate: v.Activate}
	}
	for _, v := range o.Connections {
		n.Connections = append(n.Connections, &Connection{
			In:         v.In,
			Out:        v.Out,
			Weight:     v.Weight,
			Enabled:    v.Enabled,
			Innovation: v.Innovation,
		})
	}
	return n
}

// GetActiveNodeNumber ...
func (o *Genome) GetActiveNodeNumber() int {
	return o.NextNodeID
	// nn := 0
	// for range o.Nodes {
	// 	nn++
	// }
	// return nn
}

// GetActiveConnectionNumber ...
func (o *Genome) GetActiveConnectionNumber() int {
	cn := 0
	for _, c := range o.Connections {
		if c.Enabled {
			cn++
		}
	}
	return cn
}

// Genomes ...
type Genomes []*Genome

func (s Genomes) Len() int {
	return len(s)
}

func (s Genomes) Less(i, j int) bool {
	if s[i].Fitness == s[j].Fitness {
		aid, bid := s[i].GetActiveNodeNumber(), s[j].GetActiveNodeNumber()
		if aid == bid {
			c1, c2 := s[i].GetActiveConnectionNumber(), s[j].GetActiveConnectionNumber()
			if c1 == c2 {
				return NeatRandom(0, 1) < 0.5
			}
			return c1 < c2
		}
		return aid < bid
	}
	return s[i].Fitness < s[j].Fitness
}

func (s Genomes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
