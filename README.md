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
package main

import (
	"fmt"
	"sync"
)

var msg string
var wg sync.WaitGroup

func updateMessage(s string) {
	defer wg.Done()
	msg = s
}

func main() {
	msg = "Hello, world!"

	wg.Add(2)
	go updateMessage("Hello, universe!")
	go updateMessage("Hello, cosmos!")
	wg.Wait()

	fmt.Println(msg)
}
```

```sh
go run --race .
```

### Adding mutex

```go
package main

import (
	"fmt"
	"sync"
)

var msg string
var wg sync.WaitGroup

func updateMessage(s string, m *sync.Mutex) {
	defer wg.Done()

	m.Lock()
	msg = s
	m.Unlock()
}

func main() {
	msg = "Hello, world!"

	var mutex sync.Mutex

	wg.Add(2)
	go updateMessage("Hello, universe!", &mutex)
	go updateMessage("Hello, cosmos!", &mutex)
	wg.Wait()

	fmt.Println(msg)
}
```

### Test mutex

```go
package main

import "testing"

func Test_updateMessage(t *testing.T) {
	msg = "Hello, world!"

	wg.Add(2)
	go updateMessage("x")
	go updateMessage("Goodbye, cruel world!")
	wg.Wait()

	if msg != "Goodbye, cruel world!" {
		t.Error("incorrect value in msg")
	}
}
```

```sh
go test --race
```

### More complex mutex

```go
package main

import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup

type Income struct {
	Source string
	Amount int
}

func main() {
	// variable for bank balance
	var bankBalance int
	var balance sync.Mutex

	// print out starting values
	fmt.Printf("Initial account balance: $%d.00", bankBalance)
	fmt.Println()

	// define weekly revenue
	incomes := []Income{
		{Source: "Main job", Amount: 500},
		{Source: "Gifts", Amount: 10},
		{Source: "Part time job", Amount: 50},
		{Source: "Investments", Amount: 100},
	}

	wg.Add(len(incomes))

	// loop through 52 weeks and print out how much is made; keep a running total
	for i, income := range incomes {

		go func(i int, income Income) {
			defer wg.Done()

			for week := 1; week <= 52; week++ {
				balance.Lock()
				temp := bankBalance
				temp += income.Amount
				bankBalance = temp
				balance.Unlock()

				fmt.Printf("On week %d, you earned $%d.00 from %s\n", week, income.Amount, income.Source)
			}
		}(i, income)
	}

	wg.Wait()

	// print out final balance
	fmt.Printf("Final bank balance: $%d.00", bankBalance)
	fmt.Println()
}
```

### Test for weekly income
