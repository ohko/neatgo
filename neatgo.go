package neatgo

import (
	"math"
	"math/rand"
	"time"
)

// FitnessFunction ...
type FitnessFunction func(genomes []*Genome, generation int, population *Population)

func sigmoid(x float64) float64 {
	return (1 / (1 + math.Exp(-x)))
}

// FeedForwardNetwork ...
func FeedForwardNetwork(genome *Genome, inputs []float64) []float64 {
	outputs := []float64{}

	for i := range inputs {
		genome.Nodes[i].Value = inputs[i]
	}

	// hidden
	for n := 0; n < genome.NextNodeID; n++ {
		if genome.Nodes[n].Type != NodeTypeHidden {
			continue
		}
		genome.Nodes[n].Value = 0

		for _, c := range genome.Connections {
			if c.Enabled && c.Out == n {
				genome.Nodes[c.Out].Value += genome.Nodes[c.In].Value * c.Weight
			}
		}

		genome.Nodes[n].Value = sigmoid(genome.Nodes[n].Value)
	}

	// output
	for n := 0; n < genome.NextNodeID; n++ {
		if genome.Nodes[n].Type != NodeTypeOutput {
			continue
		}
		genome.Nodes[n].Value = 0

		for _, c := range genome.Connections {
			if c.Enabled && c.Out == n {
				genome.Nodes[c.Out].Value += genome.Nodes[c.In].Value * c.Weight
			}
		}

		outputs = append(outputs, sigmoid(genome.Nodes[n].Value))
	}

	return outputs
}

var randBool = false

// NeatRandom ...
func NeatRandom(min, max float64) float64 {
	if min == 0 && max == 0 {
		return 0
	}

	if !randBool {
		randBool = true
		rand.Seed(time.Now().UnixNano())
	}
	return min + rand.Float64()*(max-min)
}

// RandIntn return min <= x <= max
// mrand "math/rand"
func RandIntn(min, max int) int {
	if min == 0 && max == 0 {
		return 0
	}
	if !randBool {
		randBool = true
		rand.Seed(time.Now().UnixNano())
	}
	return rand.Intn(max+1-min) + min
	// return mrand.New(mrand.NewSource(time.Now().UnixNano())).Intn((max-min)+1) + min
}
