# Working with Concurrency in Go

## Introduction

### Introduction

- Don't communicate by sharing memory: share memory by communicating
- Use message passing between Goroutines
- Minimum complexity

### Install go

```sh
go version
```

### Install make

```sh
brew install make
```

## Goroutines

### Creating goroutine

```sh
go mod init first-example
go run .
```

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	go print("1st thing")

	time.Sleep(time.Second)

	print("2nd thing")

}

func print(s string) {
	fmt.Println(s)
}
```

### WaitGroup

```go
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
```

## Race conditions

### Race conditions example

```go

```
