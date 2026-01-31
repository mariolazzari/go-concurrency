package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	words := []string{"alpha", "beta", "delta", "gamma", "pi", "zeta", "eta", "theta", "epsilon"}

	wg.Add(len(words))

	for i, word := range words {
		go print(fmt.Sprintf("%d: %s", i, word), &wg)
	}

	wg.Wait()

	wg.Add(1)
	print("2nd thing", &wg)

}

func print(s string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println(s)
}
