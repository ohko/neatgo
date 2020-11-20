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
		o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeOutput, Value: 0}

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
}
func (o *Genome) nextGeneration(n int, stdev float64) {
	o.mutateWeight(n, stdev)
	// r := float64(n+1) / float64(o.Population.genomeNumber)
	// r1, r2, r3, r4 := r, r, r, r
	r, r1, r2, r3, r4 := NeatRandom(0, 1), o.Population.Options.AddNode, o.Population.Options.RemoveNode, o.Population.Options.AddConnection, o.Population.Options.RemoveConnection
	if r < r1 {
		o.addNode()
	} else if r < r1+r2 {
		o.removeNode()
	} else if r < r1+r2+r3 {
		o.addConnection()
	} else if r < r1+r2+r3+r4 {
		o.removeConnection()
	}
}
func (o *Genome) mutateWeight(n int, stdev float64) {
	for i := 0; i < len(o.Connections); i++ {
		if NeatRandom(0, 1) < 0.01 {
			o.Connections[i].Weight = NeatRandom(-1, 1)
		} else if NeatRandom(0, 1) < o.Population.Options.MutateWeight {
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
			if o.Nodes[out].Type == NodeTypeInput {
				continue
			}
			if out <= in {
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
func (o *Genome) removeConnection() {
	if len(o.Connections) < o.Population.outputNumber {
		return
	}
	i := RandIntn(0, len(o.Connections)-1)
	if o.Nodes[o.Connections[i].Out].Type == NodeTypeOutput {
		count := o.getConnectionToOutput(o.Connections[i].Out)
		if count <= 1 {
			return
		}
	}
	o.Connections = append(o.Connections[:i], o.Connections[i+1:]...)
}
func (o *Genome) getConnectionToOutput(n int) int {
	count := 0
	for _, c := range o.Connections {
		if c.Out == n {
			count++
		}
	}
	return count
}
func (o *Genome) addNode() {
	outs := []*Connection{}
	for a := range o.Connections {
		if _, ok := o.Nodes[o.Connections[a].Out]; !ok {
			continue
		}
		if o.Nodes[o.Connections[a].Out].Type == NodeTypeOutput {
			outs = append(outs, o.Connections[a])
		}
	}

	o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeHidden, Value: 0}

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
func (o *Genome) removeNode() {
	removeIndex := -1
	for i := range o.Nodes {
		if o.Nodes[i].Type != NodeTypeHidden {
			continue
		}
		removeIndex = i
	}
	if removeIndex == -1 {
		return
	}

	for k := 0; k < len(o.Connections); k++ {
		if o.Connections[k].In == removeIndex || o.Connections[k].Out == removeIndex {
			if o.Nodes[o.Connections[k].Out].Type == NodeTypeOutput {
				count := o.getConnectionToOutput(o.Connections[k].Out)
				if count <= 1 {
					return
				}
			}
			o.Connections = append(o.Connections[:k], o.Connections[k+1:]...)
			k--
		}
	}

	delete(o.Nodes, removeIndex)
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
		n.Nodes[k] = &Node{Index: v.Index, Type: v.Type, Value: v.Value}
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
	nn := 0
	for range o.Nodes {
		nn++
	}
	return nn
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
