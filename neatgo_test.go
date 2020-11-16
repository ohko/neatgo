package neatgo

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"neatgo/mnist"
	"runtime"
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
		fmt.Printf("generation:%d nodes:%d connections:%d fitness:%.16f\n", generation, genomes[0].NextNodeID-1, len(genomes[0].Connections), genomes[0].Fitness)

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
	winner := pop.run(fitnessFunction, -1, "")
	// fmt.Println(winner.ToJSON())

	{ // test
		winners := winner.population.Winners
		fmt.Printf("nodes:%d connections:%d fitness:%.16f\n", winners[0].NextNodeID-1, len(winners[0].Connections), winners[0].Fitness)
		genome, _ := NewGenome(pop)
		genome.LoadJSON(winner.ToJSON())
		for _, d := range data {
			outputs := FeedForwardNetwork(genome, d["inputs"])
			fmt.Printf("%.0f => %.0f ~ %.16f\n", d["inputs"], d["outputs"], outputs)
		}
	}
}

// go test neatgo -run TestMnist -v -count=1 -timeout=1000h
func TestMnist(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Print("\033c")

	dataTrain, err := mnist.ReadTrainSet("./mnist/MNIST_data")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("MNISST train: N:%v | W:%v | H:%v", dataTrain.N, dataTrain.W, dataTrain.H)

	dataCheck, err := mnist.ReadTestSet("./mnist/MNIST_data")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("MNISST test: N:%v | W:%v | H:%v", dataCheck.N, dataCheck.W, dataCheck.H)

	getBits := func(img [][]uint8) []float64 {
		bits := make([]float64, dataTrain.W*dataTrain.H)
		pos := 0
		for _, vv := range img {
			for _, vvv := range vv {
				bits[pos] = float64(vvv) / 0xff
				// if vvv > 128 {
				// 	bits[pos] = 1
				// } else {
				// 	bits[pos] = 0
				// }
				pos++
			}
		}
		return bits
	}

	trainCount, checkCount := dataTrain.N, dataCheck.N

	dataTrainSet := [][][]float64{}
	dataCheckSet := [][][]float64{}
	for k, v := range dataTrain.Data {
		if k >= trainCount {
			break
		}
		bits := getBits(v.Image)
		want := float64(v.Digit)
		dataTrainSet = append(dataTrainSet, [][]float64{bits, {want}})
	}
	for k, v := range dataCheck.Data {
		if k >= checkCount {
			break
		}
		bits := getBits(v.Image)
		want := float64(v.Digit)
		dataCheckSet = append(dataCheckSet, [][]float64{bits, {want}})
	}

	pop, _ := NewPopulation(dataTrain.W*dataTrain.H, 10, 10, 0.9)

	outputChk := func(genome *Genome, inputs []float64, want int) bool {
		outputs := FeedForwardNetwork(genome, inputs)
		maxV, maxI := 0.0, 0
		for ok, ov := range outputs {
			if ov > maxV {
				maxV = ov
				maxI = ok
			}
		}
		if maxV > 0.8 {
			if want == maxI {
				return true
			}
		}
		return false
	}

	resultChk := func() { // check
		right := 0
		genome, _ := NewGenome(pop)
		js, err := ioutil.ReadFile("neatgo_mnist.json")
		if err != nil {
			return
		}
		genome.LoadJSON(string(js))

		for _, v := range dataCheckSet {
			if outputChk(genome, v[0], int(v[1][0])) {
				right++
			}
		}

		fmt.Printf("[check]count: %d right: %d (%.3f%%)\n", dataCheck.N, right, float64(right)/float64(checkCount)*100)
	}

	maxFitness := 0.0
	fitnessFunction := func(genomes []*Genome, generation int, population *Population) {
		fmt.Printf("generation:%d nodes:%d connections:%d fitness:%.3f%%\n", generation, genomes[0].NextNodeID-1, len(genomes[0].Connections), genomes[0].Fitness*100)
		// save
		if genomes[0].Fitness > maxFitness {
			if maxFitness != 0 {
				fmt.Println("save")
				ioutil.WriteFile("neatgo_mnist.json", []byte(genomes[0].ToJSON()), 0644)
			}
			maxFitness = genomes[0].Fitness
		}
		// check
		if generation%10 == 0 {
			resultChk()
		}

		var wg sync.WaitGroup

		for _, genome := range genomes {
			wg.Add(1)
			go func(genome *Genome) {
				right := 0
				for _, v := range dataTrainSet {
					if outputChk(genome, v[0], int(v[1][0])) {
						right++
					}
				}
				genome.Fitness = float64(right) / float64(trainCount)

				wg.Done()
			}(genome)
		}

		wg.Wait()
	}

	{ // train
		js, _ := ioutil.ReadFile("neatgo_mnist.json")
		winner := pop.run(fitnessFunction, -1, string(js))
		ioutil.WriteFile("neatgo_mnist.json", []byte(winner.ToJSON()), 0644)
		winners := winner.population.Winners
		fmt.Printf("nodes:%d connections:%d fitness:%.3f\n", winners[0].NextNodeID, len(winners[0].Connections), winners[0].Fitness)
	}

	// check
	resultChk()
}
