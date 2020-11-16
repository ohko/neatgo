package mnist

import (
	"fmt"
	"testing"
)

func Test_mnist(t *testing.T) {
	dataSet, err := ReadTrainSet("./MNIST_data")
	// or dataSet, err := ReadTestSet("./MNIST_data")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dataSet.N) // number of data
	fmt.Println(dataSet.W) // image width [pixel]
	fmt.Println(dataSet.H) // image height [pixel]
	for i := 0; i < 10; i++ {
		printData(dataSet, i)
	}
	printData(dataSet, dataSet.N-1)
}
func printData(dataSet *DataSet, index int) {
	data := dataSet.Data[index]
	fmt.Println(data.Digit) // print Digit (label)
	PrintImage(data.Image)  // print Image
}
