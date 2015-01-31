package main

import (
	"testing"

	"github.com/didiercrunch/doorman"
	"gopkg.in/mgo.v2/bson"
)

func TestDoormanDistribution(t *testing.T) {
	wab = doorman.NewDoorman(bson.NewObjectId(), []float64{0.25, 0.5, 0.10, 0.05, 0.1})
	n := 10000
	cases := make([]float64, wab.Length())
	for i := 0; i <= n; i++ {
		cases[wab.GetRandomCase()] += 1
	}
	expt := make([]float64, wab.Length())
	for i, p := range wab.Probabilities {
		expt[i] = p * float64(n)
	}
	//  Todo: use chi square test to assert the results are not absurde
	//if err := AssertTwoDistributionAreProbabliNotDifferent(expt, cases, 0.001); err != nil {
	//	t.Error(err)
	//}

}
