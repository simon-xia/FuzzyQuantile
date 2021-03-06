package fuzzyQuantile

import (
	"log"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestFuzzyQuantileBiased(t *testing.T) {
	const testDataStreamSize = 1000000
	//Logger = log.New(os.Stderr, "[FuzzyQuantile] ", log.LstdFlags)

	arr := mockDataStream(testDataStreamSize)
	fq := NewFuzzyQuantile(nil)

	startAt := time.Now()
	for i := range arr {
		fq.Insert(arr[i])
	}
	t.Logf("%d items insert takes: %+v", testDataStreamSize, time.Since(startAt))

	fq.store.flush()
	t.Log(fq.Describe())

	for i, p := range []float64{0.5, 0.8, 0.95} {
		v, er := fq.Query(p)
		if er != nil {
			t.Fatal(er)
		}
		checkResult(t, i, v, p, defaultBiasedEpsilon, testDataStreamSize)
	}
}

func TestFuzzyQuantileTarget(t *testing.T) {
	const testDataStreamSize = 10000000
	arr := mockDataStream(testDataStreamSize)
	testQuantiles := []Quantile{
		NewQuantile(0.5, 0.01),
		NewQuantile(0.8, 0.001),
		NewQuantile(0.95, 0.0001),
	}

	fq := NewFuzzyQuantile(&FuzzyQuantileConf{Quantiles: testQuantiles})
	startAt := time.Now()
	for i := range arr {
		fq.Insert(arr[i])
	}
	t.Logf("%d items insert takes: %+v", testDataStreamSize, time.Since(startAt))

	fq.store.flush()
	t.Log(fq.Describe())

	for i, q := range testQuantiles {
		v, er := fq.Query(q.quantile)
		if er != nil {
			t.Fatal(er)
		}
		checkResult(t, i, v, q.quantile, q.err, testDataStreamSize)
	}
}

func checkResult(t *testing.T, i int, v, quantile, err float64, cnt uint64) {

	ae := math.Abs((float64(cnt)*quantile - v)) / float64(cnt)
	t.Logf("test case %d result: query %f%% percentil with expected error %f: get value(%f) actual error(%f)\n", i+1, quantile*100, err, v, ae)
	if ae > err {
		t.Fatalf("test case %d failed: expect error %f, actual error %f", i+1, err, ae)
	}
}

func mockDataStream(cnt int) (arr []float64) {
	arr = make([]float64, cnt)
	for i := range arr {
		arr[i] = float64(i)
	}
	shuffle(arr, cnt)
	return
}

func shuffle(arr []float64, n int) {
	for i := 0; i < n; i++ {
		j := rand.Intn(len(arr))
		k := rand.Intn(len(arr))
		arr[j], arr[k] = arr[k], arr[j]
	}
	return
}

// This example shows biased quantile estimation
// Given a expected error which is defaultBiasedEpsilon (0.1%) as default, structure FuzzyQuantile will keep you quantile query with that error
func ExampleFuzzyQuantile_biasedQuantile() {

	fq := NewFuzzyQuantile(nil)

	// valueChan repsent a data stream source
	valueChan := make(chan float64)
	for v := range valueChan {
		fq.Insert(v)
	}
	// valueChan close at other place

	v, er := fq.Query(0.8)
	if er != nil {
		// handle error
	}
	log.Printf("success get 80th percentile value %v", v)
}

// This example show target quantile estimation
// Given a set of Quantiles, each Quantile instance repsent a pair (quantile, error) which means expected quantile value with the error
// And query will give the quantile value with corresponding error
func ExampleFuzzyQuantile_targetQuantile() {

	testQuantiles := []Quantile{
		NewQuantile(0.5, 0.01),
		NewQuantile(0.8, 0.001),
		NewQuantile(0.95, 0.0001),
	}

	fq := NewFuzzyQuantile(&FuzzyQuantileConf{Quantiles: testQuantiles})

	// valueChan repsent a data stream source
	valueChan := make(chan float64)
	for v := range valueChan {
		fq.Insert(v)
	}
	// valueChan close at other place

	v, er := fq.Query(0.8)
	if er != nil {
		// handle error
	}
	log.Printf("success get 80th percentile value %v", v)
}
