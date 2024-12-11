package kmeans

import (
	"math"
	"reflect"
	"testing"
)

func TestKMeansDeterministic(t *testing.T) {
	data := [][]float64{
		{1.0, 2.0},
		{2.0, 1.0},
		{3.0, 4.0},
		{5.0, 7.0},
		{3.5, 5.0},
		{4.5, 5.0},
		{3.5, 4.5},
	}
	k := 2
	maxIter := 10

	centroids1 := KMeans(data, k, maxIter)
	centroids2 := KMeans(data, k, maxIter)

	// Compare outputs
	for i := range centroids1 {
		for j := range centroids1[i] {
			if centroids1[i][j] != centroids2[i][j] {
				t.Errorf("Results are not deterministic: %+v != %+v", centroids1, centroids2)
			}
		}
	}
}

func TestEuclideanDistance(t *testing.T) {
	a := []float64{1.0, 2.0}
	b := []float64{4.0, 6.0}

	expected := 5.0
	result := euclideanDistance(a, b)

	if math.Abs(expected-result) > 1e-6 {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestKMeansEmptyData(t *testing.T) {
	data := [][]float64{}
	k := 2
	maxIter := 10

	centroids := KMeans(data, k, maxIter)

	if len(centroids) != 0 {
		t.Errorf("Expected 0 centroids, got %d", len(centroids))
	}
}

func TestKMeansSinglePoint(t *testing.T) {
	data := [][]float64{{1.0, 2.0}}
	k := 1
	maxIter := 10

	centroids := KMeans(data, k, maxIter)

	if len(centroids) != 1 {
		t.Errorf("Expected 1 centroid, got %d", len(centroids))
	}
	if !reflect.DeepEqual(centroids[0], data[0]) {
		t.Errorf("Expected centroid to match data point: %+v != %+v", centroids[0], data[0])
	}
}

func TestKMeansMultipleClusters(t *testing.T) {
	data := [][]float64{
		{1.0, 2.0},
		{2.0, 1.0},
		{8.0, 9.0},
		{9.0, 8.0},
		{50.0, 50.0},
	}
	k := 3
	maxIter := 10

	centroids := KMeans(data, k, maxIter)

	if len(centroids) != k {
		t.Errorf("Expected %d centroids, got %d", k, len(centroids))
	}
}

func TestKMeansConvergence(t *testing.T) {
	data := [][]float64{
		{1.0, 2.0},
		{2.0, 1.0},
		{3.0, 4.0},
		{5.0, 7.0},
		{3.5, 5.0},
		{4.5, 5.0},
		{3.5, 4.5},
	}
	k := 2
	maxIter := 100

	centroids := KMeans(data, k, maxIter)

	// Ensure centroids converge (optional)
	if len(centroids) != k {
		t.Errorf("Expected %d centroids, got %d", k, len(centroids))
	}
}

func BenchmarkKMeans(b *testing.B) {
	data := make([][]float64, 1000)
	for i := range data {
		data[i] = []float64{float64(i), float64(i * 2)}
	}
	k := 5
	maxIter := 100

	for i := 0; i < b.N; i++ {
		KMeans(data, k, maxIter)
	}
}
