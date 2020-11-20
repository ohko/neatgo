package neatgo

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

// go test neatgo -run TestXOR -v -count=1
func TestXOR(t *testing.T) {
	// fmt.Print("\033c")

	data := []map[string][]float64{
		{"inputs": {0, 0}, "outputs": {0}},
		{"inputs": {0, 1}, "outputs": {1}},
		{"inputs": {1, 0}, "outputs": {1}},
		{"inputs": {1, 1}, "outputs": {0}},
	}

	fitnessFunction := func(genomes []*Genome, generation int, population *Population) {
		if generation%10 == 0 {
			fmt.Printf("generation:%d nodes:%d/%d connections:%d/%d fitness:%.16f\n", generation, genomes[0].GetActiveNodeNumber(), genomes[0].NextNodeID, genomes[0].GetActiveConnectionNumber(), len(genomes[0].Connections), genomes[0].Fitness)
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

	pop, _ := NewPopulation(2, 1, 100, 4)
	winner := pop.Run(fitnessFunction, -1, "")
	// fmt.Println(winner.ToJSON())
	// ioutil.WriteFile("neatgo_xor.json", []byte(winner.ToJSON()), 0644)

	Visualization(winner, "visualization_xor.html")

	{ // test
		winners := winner.Population.Winners
		fmt.Printf("nodes:%d connections:%d fitness:%.16f\n", winners[0].GetActiveNodeNumber(), winners[0].GetActiveConnectionNumber(), winners[0].Fitness)
		genome, _ := NewGenome(pop)
		genome.LoadJSON(winner.ToJSON())
		for _, d := range data {
			outputs := FeedForwardNetwork(genome, d["inputs"])
			fmt.Printf("%.0f => %.0f ~ %.16f\n", d["inputs"], d["outputs"], outputs)
		}
	}
}
