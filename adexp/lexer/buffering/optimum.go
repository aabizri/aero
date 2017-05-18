package buffering

import (
	"io"
	"testing"

	"github.com/aabizri/aero/adexp/lexer/ondemand"
	"github.com/aabizri/aero/internal/repeating"
)

const benchString = " -TITLE   SAM -   ARCID AFR 456 -IFPLID XX11111111 -ADEP   LFPG  -ADES  EGLL -EOBD  140110   -EOBT 0900 -CTOT 0930 -REGUL XXXXXXX -REGCAUSE XXXX -TAXITIME XXXXX -GEO -GEOID 01 -LATTD 520000N -LONGTD 0150000W -BEGIN ADDR -FAC LLEVZPZX -FAC LFFFZQZX -END ADDR "

const (
	// DefaultBufferSize is the recommended one, it should work at ease on most modern machines
	// However, you may want to call OptimumBufferSize() to get a custom buffer size.
	DefaultBufferSize = 103
	floorBufferSize   = 30
	ceilBufferSize    = 180
	gap               = 20
)

// OptimumBufferSize runs a bunch of benchmarks and returns the optimum buffer size
// It is best used in an init step, certainly should not be ran more than once.
// You can expect it to run for ~30 seconds
func OptimumBufferSize() int {
	// TODO an algorithm that "hones in" on the optimal buffer size
	s, _ := optimum(floorBufferSize, ceilBufferSize)
	return s
}

// betterThan returns if a is better than b
func betterThan(a testing.BenchmarkResult, b testing.BenchmarkResult) bool {
	return float64(a.T)/float64(a.N) < float64(b.T)/float64(b.N)
}

// The algorithm used is quite like quisort
// 	- Take a floor and ceil buffer size, and run a benchmark on the 1/3 and 2/3 of the sizes
// 	- For the best buffer size on these, take a range from the indicator before that and then one after that and redo step 1
// 	- Stop when the difference between the floor and ceiling is <gap elements
func optimum(floor int, ceil int) (int, testing.BenchmarkResult) {
	sizes := [...]int{floor, floor + (ceil-floor)/3, ceil - (ceil-floor)/3, ceil}
	var (
		bestSizeIndex int
		bestResults   testing.BenchmarkResult
	)
	// Launch for the two sizes
	for i, size := range sizes[1 : len(sizes)-1] {
		results := testing.Benchmark(genBenchmark(size))
		if betterThan(results, bestResults) || i == 0 {
			bestSizeIndex, bestResults = i+1, results
		}
	}

	// If the gap is small enough, we return directly the best results without launching into recursion
	if ceil-floor <= gap {
		return sizes[bestSizeIndex], bestResults
	}

	// We run optimum on the surronding sizes of the best one
	size, results := optimum(sizes[bestSizeIndex-1], sizes[bestSizeIndex]+1)

	// If the next level of recursion is worse than this one, we return the one we have
	if betterThan(bestResults, results) {
		size, results = sizes[bestSizeIndex], bestResults
	}

	// DEBUG
	//fmt.Printf("Found the best size: %d : %s\n", size, results.String())

	// We return the values
	return size, results
}

func genBenchmark(s int) func(*testing.B) {
	return func(b *testing.B) {
		buf := repeating.NewStringReader(benchString)
		embedded := ondemand.New(buf)
		lexer := New(embedded, s)

		b.N = 6000000
		for i := 0; i < b.N; i++ {
			lexer.ReadLex()
		}

		lexer.Close()
		sub := embedded.(io.Closer)
		sub.Close()
	}
}
