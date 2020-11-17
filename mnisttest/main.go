package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"neatgo"
	"neatgo/mnist"
	"runtime"
	"sync"
)

var (
	t = flag.Bool("c", false, "Check")
	n = flag.Int("n", 100, "Train number")
	g = flag.Int("g", 10, "Genome number")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Lshortfile)

	if *g < 10 {
		*g = 10
	}

	jsonFile := "neatgo_mnist.json"
	pop, _ := neatgo.NewPopulation(28*28, 10, *g, 0.99)

	if *t {
		resultChk(pop, jsonFile)
		return
	}

	// fmt.Print("\033c")

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

	// trainCount := dataTrain.N
	trainCount := *n
	if trainCount > dataTrain.N {
		trainCount = dataTrain.N
	}
	if trainCount <= 0 {
		trainCount = 10
	}

	dataTrainSet := [][][]float64{}
	for k, v := range dataTrain.Data {
		if k >= trainCount {
			break
		}
		bits := getBits(v.Image)
		want := float64(v.Digit)
		dataTrainSet = append(dataTrainSet, [][]float64{bits, {want}})
	}

	maxFitness := 0.0
	fitnessFunction := func(genomes []*neatgo.Genome, generation int, population *neatgo.Population) {
		fmt.Printf("generation:%d nodes:%d/%d connections:%d/%d fitness:%.3f%%\n", generation, genomes[0].GetActiveNodeNumber(), genomes[0].NextNodeID, genomes[0].GetActiveConnectionNumber(), len(genomes[0].Connections), genomes[0].Fitness*100)
		// save
		if genomes[0].Fitness > maxFitness+0.001 {
			if maxFitness != 0 {
				ioutil.WriteFile(jsonFile, []byte(genomes[0].ToJSON()), 0644)
			}
			maxFitness = genomes[0].Fitness
		}

		var wg sync.WaitGroup

		wg.Add(len(genomes))
		for _, genome := range genomes {
			go func(genome *neatgo.Genome) {
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
		js, _ := ioutil.ReadFile(jsonFile)
		winner := pop.Run(fitnessFunction, -1, string(js))
		ioutil.WriteFile(jsonFile, []byte(winner.ToJSON()), 0644)
		winners := winner.Population.Winners
		fmt.Printf("nodes:%d connections:%d fitness:%.3f\n", winners[0].GetActiveNodeNumber(), winners[0].GetActiveConnectionNumber(), winners[0].Fitness)
	}

	// check
	resultChk(pop, jsonFile)
}

func outputChk(genome *neatgo.Genome, inputs []float64, want int) bool {
	outputs := neatgo.FeedForwardNetwork(genome, inputs)
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

func getBits(img [][]uint8) []float64 {
	bits := make([]float64, 28*28)
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

func resultChk(pop *neatgo.Population, jsonFile string) {
	dataCheck, err := mnist.ReadTestSet("./mnist/MNIST_data")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("MNISST test: N:%v | W:%v | H:%v", dataCheck.N, dataCheck.W, dataCheck.H)

	dataCheckSet := [][][]float64{}
	for k, v := range dataCheck.Data {
		if k >= dataCheck.N {
			break
		}
		bits := getBits(v.Image)
		want := float64(v.Digit)
		dataCheckSet = append(dataCheckSet, [][]float64{bits, {want}})
	}

	right := 0
	genome, _ := neatgo.NewGenome(pop)
	js, err := ioutil.ReadFile(jsonFile)
	if err != nil || len(js) == 0 {
		return
	}
	genome.LoadJSON(string(js))

	for _, v := range dataCheckSet {
		if outputChk(genome, v[0], int(v[1][0])) {
			right++
		}
	}

	fmt.Printf("[check]count: %d right: %d (%.3f%%)\n", dataCheck.N, right, float64(right)/float64(dataCheck.N)*100)
}
