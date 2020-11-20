package neatgo

import (
	"encoding/json"
	"math"
)

// Genome ...
type Genome struct {
	Population     *Population `json:"-"`
	Nodes          map[int]*Node
	Connections    []*Connection
	NextNodeID     int
	Fitness        float64
	MaxNodes       int
	MaxConnections int
}

// NewGenome ...
func NewGenome(population *Population) (*Genome, error) {
	return &Genome{
		Population: population,
		Nodes:      make(map[int]*Node),
	}, nil
}

func (o *Genome) init() {
	// o.MaxNodes = o.Population.inputNumber * 5
	// o.MaxConnections = o.Population.inputNumber * 10
	for i := 0; i < o.Population.inputNumber; i++ {
		o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeInput, Value: NeatRandom(-1, 1)}
		o.NextNodeID++
	}
	for j := 0; j < o.Population.outputNumber; j++ {
		o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeOutput, Value: 0}

		// for i := 0; i < o.Population.inputNumber; i++ {
		o.Connections = append(o.Connections, &Connection{
			// In: o.Nodes[i].Index,
			In:         RandIntn(0, o.Population.inputNumber-1),
			Out:        o.NextNodeID,
			Weight:     NeatRandom(-1, 1),
			Enabled:    true,
			Innovation: o.Population.nextInnovationID,
		})
		o.Population.nextInnovationID++
		// }

		o.NextNodeID++
	}
}
func (o *Genome) nextGeneration(n int, stdev float64) {
	o.mutateWeight(n, stdev)
	if NeatRandom(0, 1) < float64(n+1)/float64(o.Population.genomeNumber) {
		o.addNode()
	}
	// if NeatRandom(0, 1) < float64(n+1)/float64(o.Population.genomeNumber)/10 {
	// 	o.removeNode()
	// }
	if NeatRandom(0, 1) < float64(n+1)/float64(o.Population.genomeNumber) {
		o.addConnection()
	}
}
func (o *Genome) mutateWeight(n int, stdev float64) {
	for i := 0; i < len(o.Connections); i++ {
		if NeatRandom(0, 1) < 0.01 {
			o.Connections[i].Weight = NeatRandom(-1, 1)
		} else if NeatRandom(0, 1) < float64(n+1)/float64(o.Population.genomeNumber) {
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
	if o.MaxConnections > 0 && len(o.Connections) > o.MaxConnections {
		return
	}
	ins, outs := []int{}, []int{}
	for a := range o.Nodes {
		if o.Nodes[a].Type != NodeTypeOutput {
			ins = append(ins, a)
		}
		if o.Nodes[a].Type != NodeTypeInput {
			outs = append(outs, a)
		}
	}

	index := ins[RandIntn(0, len(ins)-1)]
	for _, v := range outs {
		if v <= index {
			continue
		}
		for k, c := range o.Connections {
			if c.In == index && c.Out == v {
				o.Connections[k].Enabled = true
				return
			}
		}

		cc := &Connection{
			In:         index,
			Out:        v,
			Weight:     NeatRandom(-1, 1),
			Enabled:    true,
			Innovation: o.Population.nextInnovationID,
		}
		o.Connections = append(o.Connections, cc)
		o.Population.nextInnovationID++
		return
	}
}
func (o *Genome) addNode() {
	if o.MaxNodes > 0 && (o.NextNodeID > o.MaxNodes || (o.NextNodeID >= 7 && o.NextNodeID > len(o.Connections)/2)) {
		return
	}
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
	indexs := []int{}
	for i := range o.Nodes {
		if o.Nodes[i].Type != NodeTypeHidden {
			continue
		}
		indexs = append(indexs, i)
	}
	if len(indexs) == 0 {
		return
	}
	removeIndex := indexs[RandIntn(0, len(indexs)-1)]

	ins, outs := []int{}, []int{}
	for k := 0; k < len(o.Connections); k++ {
		if o.Connections[k].Out == removeIndex {
			ins = append(ins, o.Connections[k].In)
			o.Connections = append(o.Connections[:k], o.Connections[k+1:]...)
		}
	}
	if len(ins) == 0 {
		return
	}
	for k := 0; k < len(o.Connections); k++ {
		if o.Connections[k].In == removeIndex {
			outs = append(outs, o.Connections[k].Out)
			o.Connections = append(o.Connections[:k], o.Connections[k+1:]...)
		}
	}

	for _, j := range outs {
		for _, i := range ins {
			has := false
			for k, c := range o.Connections {
				if c.In == i && c.Out == j {
					has = true
					o.Connections[k].Enabled = true
					break
				}
			}
			if has {
				continue
			}
			o.Connections = append(o.Connections, &Connection{
				In:         i,
				Out:        j,
				Weight:     NeatRandom(-1, 1),
				Enabled:    true,
				Innovation: o.Population.nextInnovationID,
			})
			o.Population.nextInnovationID++
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
	n.MaxNodes = o.MaxNodes
	n.MaxConnections = o.MaxConnections
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
	return o.NextNodeID
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
