package neatgo

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

// go test neatgo -run TestXOR -v -count=1
func TestXOR(t *testing.T) {
	fmt.Print("\033c")

	data := []map[string][]float64{
		{"inputs": {0, 0}, "outputs": {0}},
		{"inputs": {0, 1}, "outputs": {1}},
		{"inputs": {1, 0}, "outputs": {1}},
		{"inputs": {1, 1}, "outputs": {0}},
	}

	fitnessFunction := func(genomes []*Genome, generation int, population *Population) {
		if generation%10 == 0 {
			fmt.Printf("generation:%d nodes:%d connections:%d fitness:%.16f\n", generation, genomes[0].NextNodeID-1, len(genomes[0].Connections), genomes[0].Fitness)
		}

		var wg sync.WaitGroup

		for _, genome := range genomes {
			wg.Add(1)
			go func(genome *Genome) {
				genome.Fitness = 4
				for _, d := range data {
					outputs := FeedForwardNetwork(genome, d["inputs"])
					genome.Fitness -= math.Pow(outputs[0]-d["outputs"][0], 2)
				}
				wg.Done()
			}(genome)
		}

		wg.Wait()
	}

	pop, _ := NewPopulation(2, 1, 10, 4)
	winner := pop.run(fitnessFunction, -1)
	fmt.Println(winner.ToJSON())

	{ // test
		winners := winner.population.Winners
		fmt.Printf("nodes:%d connections:%d fitness:%.16f\n", winners[0].NextNodeID-1, len(winners[0].Connections), winners[0].Fitness)
		genome, _ := NewGenome(pop)
		genome.LoadJSON(winner.ToJSON())
		for _, d := range data {
			outputs := FeedForwardNetwork(genome, d["inputs"])
			fmt.Printf("%f => %f ~ %.16f\n", d["inputs"], d["outputs"], outputs)
		}
	}
}
