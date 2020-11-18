package neatgo

import (
	"math"
	"sort"
)

// Population ...
type Population struct {
	inputNumber      int
	outputNumber     int
	genomeNumber     int
	fitnessThreshold float64
	nextInnovationID int64
	Winners          Genomes

	genomes Genomes
}

// NewPopulation ...
func NewPopulation(inputNumber, outputNumber, genomeNumber int, fitnessThreshold float64) (*Population, error) {
	if genomeNumber < 5 {
		genomeNumber = 5
	}
	o := &Population{
		inputNumber:      inputNumber,
		outputNumber:     outputNumber,
		genomeNumber:     genomeNumber,
		fitnessThreshold: fitnessThreshold,
		nextInnovationID: 0,
		Winners:          []*Genome{},
	}
	return o, nil
}

// Run ...
func (o *Population) Run(fitnessFunction FitnessFunction, generations int, initJSON string) *Genome {
	o.createGenome(initJSON)
	if generations < 0 {
		generations = math.MaxInt32
	}
	for n := 0; n < generations; n++ {
		fitnessFunction(o.genomes, n, o)

		o.sortWinners(0)
		if n+1 == generations {
			break
		}
		if o.Winners[0].Fitness >= o.fitnessThreshold {
			break
		}

		o.next()
	}

	return o.Winners[0]
}
func (o *Population) createGenome(initJSON string) {
	for i := 0; i < o.genomeNumber; i++ {
		g, _ := NewGenome(o)
		if initJSON == "" {
			g.init()
		} else {
			g.LoadJSON(initJSON)
		}
		o.genomes = append(o.genomes, g)
	}
	if initJSON != "" {
		o.Winners = append(o.Winners, o.genomes[0].clone())
	}
}
func (o *Population) sortWinners(n int) {
	o.Winners = o.Winners[:n]
	sort.Sort(sort.Reverse(o.genomes))
	o.Winners = append(o.Winners, o.genomes[:4]...)
	sort.Sort(sort.Reverse(o.Winners))
	o.Winners = o.Winners[:4]
}
func (o *Population) next() {
	// sum := 0.0
	// for i := range o.genomes {
	// 	sum += math.Pow(o.fitnessThreshold-o.genomes[i].Fitness, 2)
	// }
	// stdev := math.Sqrt(sum / (float64(len(o.genomes))))
	// minStdev := stdev / o.fitnessThreshold

	var a, b *Genome
	for i := range o.genomes {
		if i < 4 {
			a = o.Winners[0].clone()
			b = o.Winners[1].clone()
			o.genomes[i] = a.crossover(b)
		} else if i < o.genomeNumber-2 {
			a = o.Winners[RandIntn(0, 3)].clone()
			b = o.Winners[RandIntn(0, 3)].clone()
			o.genomes[i] = a.crossover(b)
		} else {
			o.genomes[i] = o.Winners[RandIntn(0, 3)].clone()
		}
		o.genomes[i].nextGeneration(i, 0)
	}
}
