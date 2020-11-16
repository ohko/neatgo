package neatgo

import (
	"encoding/json"
)

// Genome ...
type Genome struct {
	population  *Population
	Nodes       map[int]*Node
	Connections []*Connection
	NextNodeID  int
	Fitness     float64
}

// NewGenome ...
func NewGenome(population *Population) (*Genome, error) {
	return &Genome{population: population}, nil
}

func (o *Genome) init() {
	o.Nodes = make(map[int]*Node)
	for i := 0; i < o.population.inputNumber; i++ {
		o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeInput, Value: NeatRandom(-1, 1)}
		o.NextNodeID++
	}
	for j := 0; j < o.population.outputNumber; j++ {
		o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeOutput, Value: 0}

		for i := 0; i < o.population.inputNumber; i++ {
			o.Connections = append(o.Connections, &Connection{
				In:         o.Nodes[i].Index,
				Out:        o.NextNodeID,
				Weight:     NeatRandom(-1, 1),
				Enabled:    true,
				Innovation: o.population.nextInnovationID,
			})
			o.population.nextInnovationID++
		}

		o.NextNodeID++
	}
}
func (o *Genome) nextGeneration(n int) {
	o.mutateWeight(n)
	if NeatRandom(0, 1) < float64(n+1)/float64(o.population.genomeNumber) {
		o.addNode()
	}
	if NeatRandom(0, 1) < float64(n+1)/float64(o.population.genomeNumber) {
		o.addConnection()
	}
}
func (o *Genome) mutateWeight(n int) {
	for i := 0; i < len(o.Connections); i++ {
		if NeatRandom(0, 1) < 0.001 {
			o.Connections[i].Weight = NeatRandom(-1, 1)
		} else if NeatRandom(0, 1) < float64(n+1)/float64(o.population.genomeNumber) {
			o.Connections[i].Weight += NeatRandom(-1, 1) * float64(n+1)
		}
	}
}
func (o *Genome) crossover(b *Genome) *Genome {
	for m := range o.Connections {
		for n := range b.Connections {
			if b.Connections[n].Innovation != o.Connections[m].Innovation {
				continue
			}

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
	found := false
	for a := range o.Nodes {
		if found {
			break
		}
		if o.Nodes[a].Type == NodeTypeOutput {
			continue
		}

		for b := range o.Nodes {
			if found {
				break
			}
			if a == b || o.Nodes[b].Type == NodeTypeInput {
				continue
			}

			has := false
			for _, c := range o.Connections {
				if (c.In == a && c.Out == b) || (c.In == b && c.Out == a) {
					has = true
					break
				}
			}
			if has {
				continue
			}

			found = true
			cc := &Connection{
				In:         a,
				Out:        b,
				Weight:     NeatRandom(-1, 1),
				Enabled:    true,
				Innovation: o.population.nextInnovationID,
			}
			if a > b {
				cc.In, cc.Out = b, a
			}
			o.Connections = append(o.Connections, cc)
			o.population.nextInnovationID++
			break
		}
	}
}
func (o *Genome) addNode() {
	for i := range o.Nodes {
		if o.Nodes[i].Type == NodeTypeInput {
			continue
		}

		o.Nodes[o.NextNodeID] = &Node{Index: o.NextNodeID, Type: NodeTypeHidden, Value: 0}

		c := RandIntn(0, len(o.Connections)-1)
		o.Connections[c].Enabled = false
		o.Connections = append(o.Connections, &Connection{
			In:         o.Connections[c].In,
			Out:        o.NextNodeID,
			Weight:     NeatRandom(-1, 1),
			Enabled:    true,
			Innovation: o.population.nextInnovationID,
		})
		o.population.nextInnovationID++
		o.Connections = append(o.Connections, &Connection{
			In:         o.NextNodeID,
			Out:        o.Connections[c].Out,
			Weight:     NeatRandom(-1, 1),
			Enabled:    true,
			Innovation: o.population.nextInnovationID,
		})
		o.population.nextInnovationID++

		o.NextNodeID++
		break
	}
}

// ToJSON ...
func (o *Genome) ToJSON() string {
	tmp := o.clone()
	innov := int64(0)
	connections := []*Connection{}
	for _, v := range tmp.Connections {
		if v.Enabled {
			v.Innovation = innov
			connections = append(connections, v)
			innov++
		}
	}
	tmp.Connections = connections
	bs, _ := json.Marshal(tmp)
	return string(bs)
}

// LoadJSON ...
func (o *Genome) LoadJSON(js string) error {
	return json.Unmarshal([]byte(js), o)
}
func (o *Genome) clone() *Genome {
	n := &Genome{
		population:  o.population,
		Nodes:       map[int]*Node{},
		Connections: []*Connection{},
		NextNodeID:  o.NextNodeID,
		Fitness:     o.Fitness,
	}
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

// Genomes ...
type Genomes []*Genome

func (s Genomes) Len() int {
	return len(s)
}

func (s Genomes) Less(i, j int) bool {
	if s[i].Fitness == s[j].Fitness {
		if s[i].NextNodeID == s[j].NextNodeID {
			if len(s[i].Connections) == len(s[j].Connections) {
				return NeatRandom(0, 1) < 0.5
			}
			return len(s[i].Connections) < len(s[j].Connections)
		}
		return s[i].NextNodeID < s[j].NextNodeID
	}
	return s[i].Fitness < s[j].Fitness
}

func (s Genomes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
