// Pruned backtracking search for B_W=1 matrices.
// For (t=3, N=10): searches {-1,0,1}^{8×3} with row-by-row early rejection.
package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"compact-the/lsss"
)

func main() {
	N := 10
	t := 3
	rows := N - t + 1 // 8
	cols := t          // 3

	fmt.Printf("Searching for B_W=1 matrix: N=%d, t=%d, W is %d×%d\n", N, t, rows, cols)
	fmt.Printf("Search space: 3^%d = ~%.0e candidates (before pruning)\n", rows*cols, 1.0)

	// Precompute all qualifying sets
	allSets := lsss.AllQualifyingSets(N, t)
	fmt.Printf("Qualifying sets: C(%d,%d) = %d\n", N, t, len(allSets))

	// Group qualifying sets by the maximum L-party index they contain.
	// After placing W rows 0..k (M rows t-1..t-1+k), we can check
	// sets whose max party index <= t-1+k.
	// setsUpTo[k] = sets whose max index <= t-1+k (i.e., only involve
	// F-parties and L-parties 0..k).
	setsUpTo := make([][]int, rows)
	for si, S := range allSets {
		maxIdx := 0
		for _, p := range S {
			if p > maxIdx {
				maxIdx = p
			}
		}
		// maxIdx ranges from t-1=2 to N-1=9
		// L-party row k corresponds to M row t-1+k, so maxIdx = t-1+k → k = maxIdx-(t-1)
		k := maxIdx - (t - 1)
		if k >= 0 && k < rows {
			setsUpTo[k] = append(setsUpTo[k], si)
		}
	}

	// newSetsAt[k] = sets that become checkable exactly when row k is placed
	// (i.e., sets whose max index is exactly t-1+k)
	newSetsAt := make([][]int, rows)
	for k := 0; k < rows; k++ {
		if k == 0 {
			newSetsAt[k] = setsUpTo[k]
		} else {
			// Sets in setsUpTo[k] but not in setsUpTo[k-1]
			prev := make(map[int]bool)
			for _, si := range setsUpTo[k-1] {
				prev[si] = true
			}
			for _, si := range setsUpTo[k] {
				if !prev[si] {
					newSetsAt[k] = append(newSetsAt[k], si)
				}
			}
		}
	}

	fmt.Printf("New sets per row: ")
	for k := 0; k < rows; k++ {
		fmt.Printf("row%d=%d ", k, len(newSetsAt[k]))
	}
	fmt.Println()

	// Enumerate all 27 possible row values
	rowValues := make([][3]int64, 0, 27)
	for a := int64(-1); a <= 1; a++ {
		for b := int64(-1); b <= 1; b++ {
			for c := int64(-1); c <= 1; c++ {
				rowValues = append(rowValues, [3]int64{a, b, c})
			}
		}
	}

	// Parallel search: split first row across goroutines
	numWorkers := runtime.NumCPU()
	if numWorkers > 27 {
		numWorkers = 27
	}
	fmt.Printf("Workers: %d\n", numWorkers)

	var found atomic.Bool
	var resultMu sync.Mutex
	var resultW [][]int64
	start := time.Now()

	var wg sync.WaitGroup
	rowCh := make(chan int, 27)
	for i := 0; i < 27; i++ {
		rowCh <- i
	}
	close(rowCh)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for firstRowIdx := range rowCh {
				if found.Load() {
					return
				}

				W := make([][]int64, rows)
				for j := range W {
					W[j] = make([]int64, cols)
				}
				W[0][0] = rowValues[firstRowIdx][0]
				W[0][1] = rowValues[firstRowIdx][1]
				W[0][2] = rowValues[firstRowIdx][2]

				M := lsss.BuildM(W, N, t)

				// Check sets for row 0
				ok := true
				for _, si := range newSetsAt[0] {
					MS := lsss.ExtractSubmatrix(M, allSets[si])
					if lsss.Determinant(MS, t) == 0 {
						ok = false
						break
					}
				}
				if !ok {
					continue
				}

				// Backtrack from row 1
				if backtrack(W, M, N, t, 1, rows, cols, rowValues[:], allSets, newSetsAt, &found) {
					resultMu.Lock()
					resultW = make([][]int64, rows)
					for j := range W {
						resultW[j] = make([]int64, cols)
						copy(resultW[j], W[j])
					}
					resultMu.Unlock()
					found.Store(true)
					return
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	if found.Load() {
		fmt.Printf("\nFOUND B_W=1 matrix in %v!\n", elapsed)
		fmt.Println("W =")
		for _, row := range resultW {
			fmt.Printf("  {%d, %d, %d},\n", row[0], row[1], row[2])
		}

		// Verify
		M := lsss.BuildM(resultW, N, t)
		allOK := true
		for _, S := range allSets {
			MS := lsss.ExtractSubmatrix(M, S)
			det := lsss.Determinant(MS, t)
			if det == 0 {
				fmt.Printf("VERIFICATION FAILED: set %v has det=0\n", S)
				allOK = false
			}
		}
		if allOK {
			fmt.Printf("Verified: all %d qualifying sets have non-zero determinant.\n", len(allSets))
		}

		// Print as Go literal
		fmt.Println("\nGo literal for HardcodedBW1:")
		fmt.Println("return [][]int64{")
		for _, row := range resultW {
			fmt.Printf("\t{%d, %d, %d},\n", row[0], row[1], row[2])
		}
		fmt.Println("}, nil")
	} else {
		fmt.Printf("\nNo B_W=1 matrix found after exhaustive search (%v).\n", elapsed)
		os.Exit(1)
	}
}

func backtrack(W [][]int64, M [][]int64, N, t, rowIdx, rows, cols int,
	rowValues [][3]int64, allSets [][]int, newSetsAt [][]int,
	found *atomic.Bool) bool {

	if found.Load() {
		return false
	}

	if rowIdx == rows {
		return true // All rows placed successfully
	}

	mRow := t - 1 + rowIdx // M row index for this W row

	for _, rv := range rowValues {
		if found.Load() {
			return false
		}

		// Place row
		W[rowIdx][0] = rv[0]
		W[rowIdx][1] = rv[1]
		W[rowIdx][2] = rv[2]

		// Update M row
		M[mRow][0] = -rv[0]
		M[mRow][1] = -rv[1]
		M[mRow][2] = -rv[2]

		// Check new qualifying sets that become active with this row
		ok := true
		for _, si := range newSetsAt[rowIdx] {
			MS := lsss.ExtractSubmatrix(M, allSets[si])
			if lsss.Determinant(MS, t) == 0 {
				ok = false
				break
			}
		}

		if ok {
			if backtrack(W, M, N, t, rowIdx+1, rows, cols, rowValues, allSets, newSetsAt, found) {
				return true
			}
		}
	}

	// Reset row (not strictly necessary but clean)
	W[rowIdx][0] = 0
	W[rowIdx][1] = 0
	W[rowIdx][2] = 0
	M[mRow][0] = 0
	M[mRow][1] = 0
	M[mRow][2] = 0

	return false
}
