package neatgo

import (
	"math"
	"sort"
)

// Population ...
type Population struct {
	inputNumber      int
	hiddenNumber     int
	outputNumber     int
	genomeNumber     int
	fitnessThreshold float64
	nextInnovationID int64
	Winners          Genomes
	Options          *Options

	genomes Genomes
}

// NewPopulation ...
func NewPopulation(inputNumber, hiddenNumber, outputNumber, genomeNumber int, fitnessThreshold float64, options *Options) (*Population, error) {
	if genomeNumber < 5 {
		genomeNumber = 5
	}
	if options == nil {
		options = DefaultOptions()
	}
	o := &Population{
		inputNumber:      inputNumber,
		hiddenNumber:     hiddenNumber,
		outputNumber:     outputNumber,
		genomeNumber:     genomeNumber,
		fitnessThreshold: fitnessThreshold,
		nextInnovationID: 0,
		Winners:          []*Genome{},
		Options:          options,
	}
	return o, nil
}

// Run ...
func (o *Population) Run(fitnessFunction FitnessFunction, generations int, initJSON string) *Genome {
	o.createGenome(initJSON)
	if generations < 0 {
		generations = math.MaxInt32
	}
	dis, last, keep := 0, 0.0, 0
	for n := 0; n < generations; n++ {
		fitnessFunction(o.genomes, n, o)

		o.sortWinners(keep)
		if n+1 == generations {
			break
		}
		if o.Winners[0].Fitness >= o.fitnessThreshold {
			break
		}

		if last < o.Winners[0].Fitness {
			keep = o.Options.KeepWinner
			dis = 0
		} else {
			dis++
		}
		if dis > o.Options.MaxDistance {
			keep = 0
		}
		last = o.Winners[0].Fitness
		o.next(dis)
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
	if len(o.Winners) > n {
		o.Winners = o.Winners[:n]
	}
	sort.Sort(sort.Reverse(o.genomes))
	o.Winners = append(o.Winners, o.genomes[:4]...)
	sort.Sort(sort.Reverse(o.Winners))
	o.Winners = o.Winners[:4]
}
func (o *Population) next(dis int) {
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
		o.genomes[i].nextGeneration(i, dis)
	}
}
