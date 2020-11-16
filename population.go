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
func (o *Population) run(fitnessFunction FitnessFunction, generations int) *Genome {
	o.createGenome()
	if generations < 0 {
		generations = math.MaxInt32
	}
	for n := 0; n < generations; n++ {
		fitnessFunction(o.genomes, n, o)

		o.sortWinners(1)
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
func (o *Population) createGenome() {
	for i := 0; i < o.genomeNumber; i++ {
		g, _ := NewGenome(o)
		g.init()
		o.genomes = append(o.genomes, g)
	}
}
func (o *Population) sortWinners(n int) {
	if len(o.Winners) > n {
		o.Winners = o.Winners[:n]
	}
	sort.Sort(sort.Reverse(o.genomes))
	o.Winners = append(o.Winners, o.genomes[:4]...)
	sort.Sort(sort.Reverse(o.Winners))
	o.Winners = o.Winners[:4]
}
func (o *Population) next() {
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
		o.genomes[i].nextGeneration(i)
	}
}
