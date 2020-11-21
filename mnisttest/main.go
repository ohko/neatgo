package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"neatgo"
	"neatgo/mnist"
	"runtime"
	"sync"
)

var (
	t = flag.Bool("c", false, "Check")
	n = flag.Int("n", 100, "Train count of every number")
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
	pop, _ := neatgo.NewPopulation(28/2*28/2, 0, 10, *g, 0.99, &neatgo.Options{
		KeepWinner:    1,
		AddNode:       0.2,
		AddConnection: 0.2,
		MutateWeight:  0.2,
		MaxDistance:   2,
		AllConnection: true,
	})

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

	trainCount := *n
	if trainCount > dataTrain.N {
		trainCount = dataTrain.N
	}
	if trainCount <= 0 {
		trainCount = 1
	}

	numbers := map[int]int{0: 0, 1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0, 7: 0, 8: 0, 9: 0}

	dataTrainSet := [][][]float64{}
	for _, v := range dataTrain.Data {
		bits := getBits(v.Image)
		want := float64(v.Digit)
		if numbers[v.Digit] >= trainCount {
			continue
		}
		numbers[v.Digit]++
		dataTrainSet = append(dataTrainSet, [][]float64{bits, {want}})

		sum := 0
		for _, n := range numbers {
			sum += n
		}

		if sum > trainCount*10 {
			break
		}
	}

	maxFitness := 0.0
	fitnessFunction := func(genomes []*neatgo.Genome, generation int, population *neatgo.Population) {
		fmt.Printf("generation:%d nodes:%d/%d connections:%d/%d fitness:%.3f%%\r", generation, genomes[0].GetActiveNodeNumber(), genomes[0].NextNodeID, genomes[0].GetActiveConnectionNumber(), len(genomes[0].Connections), genomes[0].Fitness*100)
		if generation%100 == 0 {
			fmt.Println()
		}
		// save
		if genomes[0].Fitness > maxFitness+0.01 {
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
				genome.Fitness = float64(right) / float64(trainCount*10)

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
	if maxV > 0.9 {
		if want == maxI {
			return true
		}
	}
	return false
}

func small(img [][]uint8) [][]uint8 {
	out := [][]uint8{}
	w, h := 2, 2
	for y := 0; y < len(img); y += h {
		row := []uint8{}
		for x := 0; x < len(img[0]); x += w {
			dot := uint8(0)
			sum := 0.0
			for a := 0; a < h; a++ {
				for b := 0; b < w; b++ {
					sum += math.Pow(float64(img[y+a][x+b]), 2)
				}
			}
			dot = uint8(math.Sqrt(sum / float64(w*h)))
			row = append(row, dot)
		}
		out = append(out, row)
	}
	return out
}

func preview(img [][]uint8) {
	for y := 0; y < len(img); y++ {
		for x := 0; x < len(img[0]); x++ {
			if img[y][x] > 128 {
				fmt.Print("1")
			} else {
				fmt.Print("0")
			}
		}
		fmt.Println()
	}
}

func getBits(img [][]uint8) []float64 {
	bits := make([]float64, 28/2*28/2)
	pos := 0
	img = small(img)
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

	checkCount := *n
	if checkCount > dataCheck.N {
		checkCount = dataCheck.N
	}
	if checkCount <= 0 {
		checkCount = 1
	}

	dataCheckSet := [][][]float64{}
	for k, v := range dataCheck.Data {
		if k >= checkCount {
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

	fmt.Printf("[check]fitness: %.2f%% right: %d/%d (%.3f%%)\n", genome.Fitness*100, right, len(dataCheckSet), float64(right)/float64(len(dataCheckSet))*100)
}
