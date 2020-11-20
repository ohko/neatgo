package neatgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"strings"
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
		if _, ok := genome.Nodes[n]; !ok {
			continue
		}
		if genome.Nodes[n].Type != NodeTypeHidden {
			continue
		}
		genome.Nodes[n].Value = 0

		for _, c := range genome.Connections {
			if c.Enabled && c.Out == n {
				if _, ok := genome.Nodes[c.In]; !ok {
					continue
				}
				if _, ok := genome.Nodes[c.Out]; !ok {
					continue
				}
				if genome.Nodes[c.In] == nil || genome.Nodes[c.Out] == nil {
					fmt.Println(genome.Nodes[c.In])
					fmt.Println(genome.Nodes[c.Out])
				}
				genome.Nodes[c.Out].Value += genome.Nodes[c.In].Value * c.Weight
			}
		}

		genome.Nodes[n].Value = sigmoid(genome.Nodes[n].Value)
	}

	// output
	for n := 0; n < genome.NextNodeID; n++ {
		if _, ok := genome.Nodes[n]; !ok {
			continue
		}
		if genome.Nodes[n].Type != NodeTypeOutput {
			continue
		}
		genome.Nodes[n].Value = 0

		for _, c := range genome.Connections {
			if c.Enabled && c.Out == n {
				if _, ok := genome.Nodes[c.In]; !ok {
					continue
				}
				if _, ok := genome.Nodes[c.Out]; !ok {
					continue
				}
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

// Visualization ...
func Visualization(genome *Genome, file string) {
	const vTpl = `
<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="echarts.min.js"></script>
</head>

<body>
	<div id="main" style="border:1px solid #CCC;width: 800px;height:800px;"></div>
	<div>{JSON}</div>
    <script type="text/javascript">
        var myChart = echarts.init(document.getElementById('main'));

        var option = {
            series: [{
                type: 'graph', layout: 'force', animation: false, label: { show: true }, edgeSymbol: ['', 'arrow'], force: { edgeLength: 100, repulsion: 1000 },
                data: {DATA},
                edges: {EDGES},
            }]
        };

        myChart.setOption(option);
    </script>
</body>

</html>`
	itemColors := map[string]string{NodeTypeInput: "red", NodeTypeHidden: "pink", NodeTypeOutput: "blue"}
	lineColors := map[bool]string{true: "black", false: "gray"}
	lineWidths := map[bool]int{true: 2, false: 1}
	datas, edges := []interface{}{}, []interface{}{}
	for k, v := range genome.Nodes {
		itemStyle := map[string]interface{}{"color": itemColors[v.Type]}
		datas = append(datas, map[string]interface{}{"name": strconv.Itoa(k), "symbolSize": 20, "draggable": true, "itemStyle": itemStyle})
	}
	for _, v := range genome.Connections {
		lineStyle := map[string]interface{}{"color": lineColors[v.Enabled], "width": lineWidths[v.Enabled]}
		edges = append(edges, map[string]interface{}{"source": strconv.Itoa(v.In), "target": strconv.Itoa(v.Out), "lineStyle": lineStyle})
	}
	bs, _ := json.Marshal(datas)
	html := strings.Replace(vTpl, "{DATA}", string(bs), 1)
	bs, _ = json.Marshal(edges)
	html = strings.Replace(html, "{EDGES}", string(bs), 1)
	html = strings.Replace(html, "{JSON}", genome.ToJSON(), 1)
	ioutil.WriteFile(file, []byte(html), 0644)
}
