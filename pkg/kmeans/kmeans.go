package kmeans

import (
	"fmt"
	"math"
	"math/rand"
)

func KMeans(data [][]float64, k int, maxIter int) [][]float64 {
	if len(data) == 0 || k <= 0 {
		fmt.Println("Invalid input: data is empty or k is non-positive.")
		return [][]float64{}
	}

	rand.Seed(42) // Fixed seed for determinism
	centroids := initializeCentroids(data, k)
	fmt.Printf("Initial centroids: %+v\n", centroids)

	for i := 0; i < maxIter; i++ {
		clusters := assignClusters(data, centroids)
		//fmt.Printf("Iteration %d: Clusters: %+v\n", i+1, clusters)
		centroids = recalculateCentroids(data, clusters, k)
		//fmt.Printf("Iteration %d: Updated centroids: %+v\n", i+1, centroids)
	}
	//fmt.Println("K-Means clustering completed.")
	return centroids
}

// initializeCentroids selects k random initial centroids from the data
func initializeCentroids(data [][]float64, k int) [][]float64 {
	if len(data) == 0 {
		return [][]float64{} // Return an empty slice if the dataset is empty
	}

	centroids := make([][]float64, k)
	for i := 0; i < k; i++ {
		centroids[i] = data[rand.Intn(len(data))]
	}
	return centroids
}

// assignClusters assigns each point in the data to the closest centroid
func assignClusters(data [][]float64, centroids [][]float64) []int {
	clusters := make([]int, len(data))
	for i, point := range data {
		clusters[i] = closestCentroid(point, centroids)
	}
	return clusters
}

// recalculateCentroids calculates new centroids as the mean of points in each cluster
func recalculateCentroids(data [][]float64, clusters []int, k int) [][]float64 {
	centroids := make([][]float64, k)
	counts := make([]int, k)

	// Initialize centroids with zero vectors
	for i := range centroids {
		centroids[i] = make([]float64, len(data[0]))
	}

	// Sum up all points in each cluster
	for i, cluster := range clusters {
		for j := range data[i] {
			centroids[cluster][j] += data[i][j]
		}
		counts[cluster]++
	}

	// Divide by the number of points in each cluster to get the mean
	for i := range centroids {
		if counts[i] == 0 {
			// Handle empty clusters by reinitializing centroids randomly
			centroids[i] = data[rand.Intn(len(data))]
		} else {
			for j := range centroids[i] {
				centroids[i][j] /= float64(counts[i])
			}
		}
	}

	return centroids
}

// closestCentroid finds the index of the centroid closest to the given point
func closestCentroid(point []float64, centroids [][]float64) int {
	minDist := math.MaxFloat64
	closest := 0

	for i, centroid := range centroids {
		dist := euclideanDistance(point, centroid)
		if dist < minDist {
			minDist = dist
			closest = i
		}
	}
	return closest
}

// euclideanDistance calculates the Euclidean distance between two points
func euclideanDistance(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += math.Pow(a[i]-b[i], 2)
	}
	return math.Sqrt(sum)
}
