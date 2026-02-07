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

```go
package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func Test_main(t *testing.T) {
	stdOut := os.Stdout
	r, w, _ := os.Pipe()

	os.Stdout = w

	main()

	_ = w.Close()

	result, _ := io.ReadAll(r)
	output := string(result)

	os.Stdout = stdOut

	if !strings.Contains(output, "$34320.00") {
		t.Error("wrong balance returned")
	}

}
```

### Producer/Consumer: useing channels

[WikiPedia](https://en.wikipedia.org/wiki/Producer%E2%80%93consumer_problem)

```go
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

const NumberOfPizzas = 10

var pizzasMade, pizzasFailed, total int

// Producer is a type for structs that holds two channels: one for pizzas, with all
// information for a given pizza order including whether it was made
// successfully, and another to handle end of processing (when we quit the channel)
type Producer struct {
	data chan PizzaOrder
	quit chan chan error
}

// PizzaOrder is a type for structs that describes a given pizza order. It has the order
// number, a message indicating what happened to the order, and a boolean
// indicating if the order was successfully completed.
type PizzaOrder struct {
	pizzaNumber int
	message     string
	success     bool
}

// Close is simply a method of closing the channel when we are done with it (i.e.
// something is pushed to the quit channel)
func (p *Producer) Close() error {
	ch := make(chan error)
	p.quit <- ch
	return <-ch
}

// makePizza attempts to make a pizza. We generate a random number from 1-12,
// and put in two cases where we can't make the pizza in time. Otherwise,
// we make the pizza without issue. To make things interesting, each pizza
// will take a different length of time to produce (some pizzas are harder than others).
func makePizza(pizzaNumber int) *PizzaOrder {
	pizzaNumber++
	if pizzaNumber <= NumberOfPizzas {
		delay := rand.Intn(5) + 1
		fmt.Printf("Received order #%d!\n", pizzaNumber)

		rnd := rand.Intn(12) + 1
		msg := ""
		success := false

		if rnd < 5 {
			pizzasFailed++
		} else {
			pizzasMade++
		}
		total++

		fmt.Printf("Making pizza #%d. It will take %d seconds....\n", pizzaNumber, delay)
		// delay for a bit
		time.Sleep(time.Duration(delay) * time.Second)

		if rnd <= 2 {
			msg = fmt.Sprintf("*** We ran out of ingredients for pizza #%d!", pizzaNumber)
		} else if rnd <= 4 {
			msg = fmt.Sprintf("*** The cook quit while making pizza #%d!", pizzaNumber)
		} else {
			success = true
			msg = fmt.Sprintf("Pizza order #%d is ready!", pizzaNumber)
		}

		p := PizzaOrder{
			pizzaNumber: pizzaNumber,
			message:     msg,
			success:     success,
		}

		return &p

	}

	return &PizzaOrder{
		pizzaNumber: pizzaNumber,
	}
}

// pizzeria is a goroutine that runs in the background and
// calls makePizza to try to make one order each time it iterates through
// the for loop. It executes until it receives something on the quit
// channel. The quit channel does not receive anything until the consumer
// sends it (when the number of orders is greater than or equal to the
// constant NumberOfPizzas).
func pizzeria(pizzaMaker *Producer) {
	// keep track of which pizza we are making
	var i = 0

	// this loop will continue to execute, trying to make pizzas,
	// until the quit channel receives something.
	for {
		currentPizza := makePizza(i)
		if currentPizza != nil {
			i = currentPizza.pizzaNumber
			select {
			// we tried to make a pizza (we send something to the data channel -- a chan PizzaOrder)
			case pizzaMaker.data <- *currentPizza:

			// we want to quit, so send pizzMaker.quit to the quitChan (a chan error)
			case quitChan := <-pizzaMaker.quit:
				// close channels
				close(pizzaMaker.data)
				close(quitChan)
				return
			}
		}
	}
}

func main() {
	// seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// print out a message
	color.Cyan("The Pizzeria is open for business!")
	color.Cyan("----------------------------------")

	// create a producer
	pizzaJob := &Producer{
		data: make(chan PizzaOrder),
		quit: make(chan chan error),
	}

	// run the producer in the background
	go pizzeria(pizzaJob)

	// create and run consumer
	for i := range pizzaJob.data {
		if i.pizzaNumber <= NumberOfPizzas {
			if i.success {
				color.Green(i.message)
				color.Green("Order #%d is out for delivery!", i.pizzaNumber)
			} else {
				color.Red(i.message)
				color.Red("The customer is really mad!")
			}
		} else {
			color.Cyan("Done making pizzas...")
			err := pizzaJob.Close()
			if err != nil {
				color.Red("*** Error closing channel!", err)
			}
		}
	}

	// print out the ending message
	color.Cyan("-----------------")
	color.Cyan("Done for the day.")

	color.Cyan("We made %d pizzas, but failed to make %d, with %d attempts in total.", pizzasMade, pizzasFailed, total)

	switch {
	case pizzasFailed > 9:
		color.Red("It was an awful day...")
	case pizzasFailed >= 6:
		color.Red("It was not a very good day...")
	case pizzasFailed >= 4:
		color.Yellow("It was an okay day....")
	case pizzasFailed >= 2:
		color.Yellow("It was a pretty good day!")
	default:
		color.Green("It was a great day!")
	}
}
```

## Dining philosophers

### The problem

[WikiPedia](https://en.wikipedia.org/wiki/Dining_philosophers_problem)

### Implementing logic

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// The Dining Philosophers problem is well known in computer science circles.
// Five philosophers, numbered from 0 through 4, live in a house where the
// table is laid for them; each philosopher has their own place at the table.
// Their only difficulty – besides those of philosophy – is that the dish
// served is a very difficult kind of spaghetti which has to be eaten with
// two forks. There are two forks next to each plate, so that presents no
// difficulty. As a consequence, however, this means that no two neighbours
// may be eating simultaneously, since there are five philosophers and five forks.
//
// This is a simple implementation of Dijkstra's solution to the "Dining
// Philosophers" dilemma.

// Philosopher is a struct which stores information about a philosopher.
type Philosopher struct {
	name      string
	rightFork int
	leftFork  int
}

// philosophers is list of all philosophers.
var philosophers = []Philosopher{
	{name: "Plato", leftFork: 4, rightFork: 0},
	{name: "Socrates", leftFork: 0, rightFork: 1},
	{name: "Aristotle", leftFork: 1, rightFork: 2},
	{name: "Pascal", leftFork: 2, rightFork: 3},
	{name: "Locke", leftFork: 3, rightFork: 4},
}

// Define a few variables.
var hunger = 3                  // how many times a philosopher eats
var eatTime = 1 * time.Second   // how long it takes to eatTime
var thinkTime = 3 * time.Second // how long a philosopher thinks
var sleepTime = 1 * time.Second // how long to wait when printing things out

func main() {
	// print out a welcome message
	fmt.Println("Dining Philosophers Problem")
	fmt.Println("---------------------------")
	fmt.Println("The table is empty.")

	// start the meal
	dine()

	// print out finished message
	fmt.Println("The table is empty.")

}

func dine() {
	eatTime = 0 * time.Second
	sleepTime = 0 * time.Second
	thinkTime = 0 * time.Second

	// wg is the WaitGroup that keeps track of how many philosophers are still at the table. When
	// it reaches zero, everyone is finished eating and has left. We add 5 (the number of philosophers) to this
	// wait group.
	wg := &sync.WaitGroup{}
	wg.Add(len(philosophers))

	// We want everyone to be seated before they start eating, so create a WaitGroup for that, and set it to 5.
	seated := &sync.WaitGroup{}
	seated.Add(len(philosophers))

	// forks is a map of all 5 forks. Forks are assigned using the fields leftFork and rightFork in the Philosopher
	// type. Each fork, then, can be found using the index (an integer), and each fork has a unique mutex.
	var forks = make(map[int]*sync.Mutex)
	for i := 0; i < len(philosophers); i++ {
		forks[i] = &sync.Mutex{}
	}

	// Start the meal by iterating through our slice of Philosophers.
	for i := 0; i < len(philosophers); i++ {
		// fire off a goroutine for the current philosopher
		go diningProblem(philosophers[i], wg, forks, seated)
	}

	// Wait for the philosophers to finish. This blocks until the wait group is 0.
	wg.Wait()
}

// diningProblem is the function fired off as a goroutine for each of our philosophers. It takes one
// philosopher, our WaitGroup to determine when everyone is done, a map containing the mutexes for every
// fork on the table, and a WaitGroup used to pause execution of every instance of this goroutine
// until everyone is seated at the table.
func diningProblem(philosopher Philosopher, wg *sync.WaitGroup, forks map[int]*sync.Mutex, seated *sync.WaitGroup) {
	defer wg.Done()

	// seat the philosopher at the table
	fmt.Printf("%s is seated at the table.\n", philosopher.name)

	// Decrement the seated WaitGroup by one.
	seated.Done()

	// Wait until everyone is seated.
	seated.Wait()

	// Have this philosopher eatTime and thinkTime "hunger" times (3).
	for i := hunger; i > 0; i-- {
		// Get a lock on the left and right forks. We have to choose the lower numbered fork first in order
		// to avoid a logical race condition, which is not detected by the -race flag in tests; if we don't do this,
		// we have the potential for a deadlock, since two philosophers will wait endlessly for the same fork.
		// Note that the goroutine will block (pause) until it gets a lock on both the right and left forks.
		if philosopher.leftFork > philosopher.rightFork {
			forks[philosopher.rightFork].Lock()
			fmt.Printf("\t%s takes the right fork.\n", philosopher.name)
			forks[philosopher.leftFork].Lock()
			fmt.Printf("\t%s takes the left fork.\n", philosopher.name)
		} else {
			forks[philosopher.leftFork].Lock()
			fmt.Printf("\t%s takes the left fork.\n", philosopher.name)
			forks[philosopher.rightFork].Lock()
			fmt.Printf("\t%s takes the right fork.\n", philosopher.name)
		}

		// By the time we get to this line, the philosopher has a lock (mutex) on both forks.
		fmt.Printf("\t%s has both forks and is eating.\n", philosopher.name)
		time.Sleep(eatTime)

		// The philosopher starts to think, but does not drop the forks yet.
		fmt.Printf("\t%s is thinking.\n", philosopher.name)
		time.Sleep(thinkTime)

		// Unlock the mutexes for both forks.
		forks[philosopher.leftFork].Unlock()
		forks[philosopher.rightFork].Unlock()

		fmt.Printf("\t%s put down the forks.\n", philosopher.name)
	}

	// The philosopher has finished eating, so print out a message.
	fmt.Println(philosopher.name, "is satisified.")
	fmt.Println(philosopher.name, "left the table.")
}
```

### Exercise: finish order

```go
package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// The Dining Philosophers problem is well known in computer science circles.
// Five philosophers, numbered from 0 through 4, live in a house where the
// table is laid for them; each philosopher has their own place at the table.
// Their only difficulty – besides those of philosophy – is that the dish
// served is a very difficult kind of spaghetti which has to be eaten with
// two forks. There are two forks next to each plate, so that presents no
// difficulty. As a consequence, however, this means that no two neighbours
// may be eating simultaneously, since there are five philosophers and five forks.
//
// This is a simple implementation of Dijkstra's solution to the "Dining
// Philosophers" dilemma.

// Philosopher is a struct which stores information about a philosopher.
type Philosopher struct {
	name      string
	rightFork int
	leftFork  int
}

// philosophers is list of all philosophers.
var philosophers = []Philosopher{
	{name: "Plato", leftFork: 4, rightFork: 0},
	{name: "Socrates", leftFork: 0, rightFork: 1},
	{name: "Aristotle", leftFork: 1, rightFork: 2},
	{name: "Pascal", leftFork: 2, rightFork: 3},
	{name: "Locke", leftFork: 3, rightFork: 4},
}

// Define a few variables.
var hunger = 3                  // how many times a philosopher eats
var eatTime = 1 * time.Second   // how long it takes to eatTime
var thinkTime = 3 * time.Second // how long a philosopher thinks
var sleepTime = 1 * time.Second // how long to wait when printing things out

// *** added this
var orderMutex sync.Mutex       // a mutex for the slice orderFinished; part of challenge!
var orderFinished []string      // the order in which philosophers finish dining and leave; part of challenge!

func main() {
	// print out a welcome message
	fmt.Println("Dining Philosophers Problem")
	fmt.Println("---------------------------")
	fmt.Println("The table is empty.")

	// *** added this
	time.Sleep(sleepTime)

	// start the meal
	dine()

	// print out finished message
	fmt.Println("The table is empty.")

	// *** added this
	time.Sleep(sleepTime)
	fmt.Printf("Order finished: %s.\n", strings.Join(orderFinished, ", "))

}

func dine() {
	eatTime = 0 * time.Second
	sleepTime = 0 * time.Second
	thinkTime = 0 * time.Second

	// wg is the WaitGroup that keeps track of how many philosophers are still at the table. When
	// it reaches zero, everyone is finished eating and has left. We add 5 (the number of philosophers) to this
	// wait group.
	wg := &sync.WaitGroup{}
	wg.Add(len(philosophers))

	// We want everyone to be seated before they start eating, so create a WaitGroup for that, and set it to 5.
	seated := &sync.WaitGroup{}
	seated.Add(len(philosophers))

	// forks is a map of all 5 forks. Forks are assigned using the fields leftFork and rightFork in the Philosopher
	// type. Each fork, then, can be found using the index (an integer), and each fork has a unique mutex.
	var forks = make(map[int]*sync.Mutex)
	for i := 0; i < len(philosophers); i++ {
		forks[i] = &sync.Mutex{}
	}

	// Start the meal by iterating through our slice of Philosophers.
	for i := 0; i < len(philosophers); i++ {
		// fire off a goroutine for the current philosopher
		go diningProblem(philosophers[i], wg, forks, seated)
	}

	// Wait for the philosophers to finish. This blocks until the wait group is 0.
	wg.Wait()
}

// diningProblem is the function fired off as a goroutine for each of our philosophers. It takes one
// philosopher, our WaitGroup to determine when everyone is done, a map containing the mutexes for every
// fork on the table, and a WaitGroup used to pause execution of every instance of this goroutine
// until everyone is seated at the table.
func diningProblem(philosopher Philosopher, wg *sync.WaitGroup, forks map[int]*sync.Mutex, seated *sync.WaitGroup) {
	defer wg.Done()

	// seat the philosopher at the table
	fmt.Printf("%s is seated at the table.\n", philosopher.name)

	// Decrement the seated WaitGroup by one.
	seated.Done()

	// Wait until everyone is seated.
	seated.Wait()

	// Have this philosopher eatTime and thinkTime "hunger" times (3).
	for i := hunger; i > 0; i-- {
		// Get a lock on the left and right forks. We have to choose the lower numbered fork first in order
		// to avoid a logical race condition, which is not detected by the -race flag in tests; if we don't do this,
		// we have the potential for a deadlock, since two philosophers will wait endlessly for the same fork.
		// Note that the goroutine will block (pause) until it gets a lock on both the right and left forks.
		if philosopher.leftFork > philosopher.rightFork {
			forks[philosopher.rightFork].Lock()
			fmt.Printf("\t%s takes the right fork.\n", philosopher.name)
			forks[philosopher.leftFork].Lock()
			fmt.Printf("\t%s takes the left fork.\n", philosopher.name)
		} else {
			forks[philosopher.leftFork].Lock()
			fmt.Printf("\t%s takes the left fork.\n", philosopher.name)
			forks[philosopher.rightFork].Lock()
			fmt.Printf("\t%s takes the right fork.\n", philosopher.name)
		}

		// By the time we get to this line, the philosopher has a lock (mutex) on both forks.
		fmt.Printf("\t%s has both forks and is eating.\n", philosopher.name)
		time.Sleep(eatTime)

		// The philosopher starts to think, but does not drop the forks yet.
		fmt.Printf("\t%s is thinking.\n", philosopher.name)
		time.Sleep(thinkTime)

		// Unlock the mutexes for both forks.
		forks[philosopher.leftFork].Unlock()
		forks[philosopher.rightFork].Unlock()

		fmt.Printf("\t%s put down the forks.\n", philosopher.name)
	}

	// The philosopher has finished eating, so print out a message.
	fmt.Println(philosopher.name, "is satisified.")
	fmt.Println(philosopher.name, "left the table.")

	// *** added this
	orderMutex.Lock()
	orderFinished = append(orderFinished, philosopher.name)
	orderMutex.Unlock()
}
```

### Dining phylosophers test

```go
package main

import (
	"testing"
	"time"
)

func Test_dine(t *testing.T) {
	eatTime = 0 * time.Second
	sleepTime = 0 * time.Second
	thinkTime = 0 * time.Second

	for i := 0; i < 10; i++ {
		orderFinished = []string{}
		dine()
		if len(orderFinished) != 5 {
			t.Errorf("incorrect length of slice; expected 5 but got %d", len(orderFinished))
		}
	}
}

func Test_dineWithVaryingDelays(t *testing.T) {
	var theTests = []struct{
		name string
		delay time.Duration
	}{
		{"zero delay", time.Second * 0},
		{"quarter second delay", time.Millisecond * 250},
		{"half second delay", time.Millisecond * 500},
	}

	for _, e := range theTests {
		orderFinished = []string{}

		eatTime = e.delay
		sleepTime = e.delay
		thinkTime = e.delay

		dine()
		if len(orderFinished) != 5 {
			t.Errorf("%s: incorrect length of slice; expected 5 but got %d", e.name, len(orderFinished))
		}
	}
}
```

## Channels

- Communications between goroutines
- Buffered or unbuffered
- Close channel once its job completes
- Fully typed

### Sleeping barber problem

[WikiPedia](https://en.wikipedia.org/wiki/Sleeping_barber_problem)

### Channels introduction

```go

```

## Channels

### Introduction to channels

```go
package main

import (
	"fmt"
	"strings"
)

// shout has two parameters: a receive only chan ping, and a send only chan pong.
// Note the use of <- in function signature. It simply takes whatever
// string it gets from the ping channel,  converts it to uppercase and
// appends a few exclamation marks, and then sends the transformed text to the pong channel.
func shout(ping <-chan string, pong chan<- string) {
	for {
		// read from the ping channel. Note that the GoRoutine waits here -- it blocks until
		// something is received on this channel.
		s := <-ping

		pong <- fmt.Sprintf("%s!!!", strings.ToUpper(s))
	}
}

func main() {
	// create two channels. Ping is what we send to, and pong is what comes back.
	ping := make(chan string)
	pong := make(chan string)

	// start a goroutine
	go shout(ping, pong)

	fmt.Println("Type something and press ENTER (enter Q to quit)")

	for {
		// print a prompt
		fmt.Print("-> ")

		// get user input
		var userInput string
		_, _ = fmt.Scanln(&userInput)

		if userInput == strings.ToLower("q") {
			// jump out of for loop
			break
		}

		// send userInput to "ping" channel
		ping <- userInput

		// wait for a response from the pong channel. Again, program
		// blocks (pauses) until it receives something from
		// that channel.
		response := <-pong

		// print the response to the console.
		fmt.Println("Response:", response)
	}

	fmt.Println("All done. Closing channels.")

	// close the channels
	close(ping)
	close(pong)
}
```

### Select statement

```go
package main

import (
	"fmt"
	"time"
)

func server1(ch chan string) {
	for {
		time.Sleep(6 * time.Second)
		ch <- "This is from server 1"
	}
}

func server2(ch chan string) {
	for {
		time.Sleep(3 * time.Second)
		ch <- "This is from server 2"
	}
}

func main() {
	fmt.Println("Select with channels")
	fmt.Println("--------------------")

	channel1 := make(chan string)
	channel2 := make(chan string)

	go server1(channel1)
	go server2(channel2)

	for {
		select {
		// because we have multiple cases listening to
		// the same channels, random ones are selected
		case s1 := <-channel1:
			fmt.Println("Case one:", s1)
		case s2 := <-channel1:
			fmt.Println("Case two:", s2)
		case s3 := <-channel2:
			fmt.Println("Case three:", s3)
		case s4 := <-channel2:
			fmt.Println("Case four:", s4)
			// default:
			// avoiding deadlock
		}
	}

}
```

### Buffered channels

```go
package main

import (
	"fmt"
	"time"
)

func listenToChan(ch chan int) {
	for {
		// print a got data message
		i := <-ch
		fmt.Println("Got", i, "from channel")

		// simulate doing a lot of work
		time.Sleep(1 * time.Second)
	}
}

func main() {
	ch := make(chan int, 10)

	go listenToChan(ch)

	for i := 0; i <= 100; i++ {
		// the first 10 times through this loop, things go quickly; after that, things slow down.
		fmt.Println("sending", i, "to channel...")
		ch <- i
		fmt.Println("sent", i, "to channel!")
	}

	fmt.Println("Done!")
	close(ch)
}
```

### Sleeping barber

```go
// This is a simple demonstration of how to solve the Sleeping Barber dilemma, a classic computer science problem
// which illustrates the complexities that arise when there are multiple operating system processes. Here, we have
// a finite number of barbers, a finite number of seats in a waiting room, a fixed length of time the barbershop is
// open, and clients arriving at (roughly) regular intervals. When a barber has nothing to do, he or she checks the
// waiting room for new clients, and if one or more is there, a haircut takes place. Otherwise, the barber goes to
// sleep until a new client arrives. So the rules are as follows:
//
//   - if there are no customers, the barber falls asleep in the chair
//   - a customer must wake the barber if he is asleep
//   - if a customer arrives while the barber is working, the customer leaves if all chairs are occupied and
//     sits in an empty chair if it's available
//   - when the barber finishes a haircut, he inspects the waiting room to see if there are any waiting customers
//     and falls asleep if there are none
//   - shop can stop accepting new clients at closing time, but the barbers cannot leave until the waiting room is
//     empty
//   - after the shop is closed and there are no clients left in the waiting area, the barber
//     goes home
//
// The Sleeping Barber was originally proposed in 1965 by computer science pioneer Edsger Dijkstra.
//
// The point of this problem, and its solution, was to make it clear that in a lot of cases, the use of
// semaphores (mutexes) is not needed.
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

// variables
var seatingCapacity = 10
var arrivalRate = 100
var cutDuration = 1000 * time.Millisecond
var timeOpen = 10 * time.Second

func main() {
	// print welcome message
	color.Yellow("The Sleeping Barber Problem")
	color.Yellow("---------------------------")

	// create channels if we need any
	clientChan := make(chan string, seatingCapacity)
	doneChan := make(chan bool)

	// create the barbershop
	shop := BarberShop{
		ShopCapacity:    seatingCapacity,
		HairCutDuration: cutDuration,
		NumberOfBarbers: 0,
		ClientsChan:     clientChan,
		BarbersDoneChan: doneChan,
		Open:            true,
	}

	color.Green("The shop is open for the day!")

	// add barbers
	shop.addBarber("Frank")

	// start the barbershop as a goroutine
	shopClosing := make(chan bool)
	closed := make(chan bool)

	go func() {
		<-time.After(timeOpen)
		shopClosing <- true
		shop.closeShopForDay()
		closed <- true
	}()

	// add clients
	i := 1

	go func() {
		for {
			// get a random number with average arrival rate
			randomMillseconds := rand.Int() % (2 * arrivalRate)
			select {
			case <-shopClosing:
				return
			case <-time.After(time.Millisecond * time.Duration(randomMillseconds)):
				shop.addClient(fmt.Sprintf("Client #%d", i))
				i++
			}
		}
	}()

	// block until the barbershop is closed
	<-closed
}
```

## Final project: subscription system

### Web app

```sh
go mod init github.com/mariolazzari/go-concurrency
go get github.com/jackc/pgx/v5
go get github.com/alexedwards/scs/v2
go get github.com/alexedwards/scs/redisstore
go get github.com/go-chi/chi/v5
go get github.com/go-chi/chi/v5/middleware
```

### Docker compose

```yaml
version: "3"

services:
  #  start Postgres, and ensure that data is stored to a mounted volume
  postgres:
    image: "postgres:14.2"
    ports:
      - "5432:5432"
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: concurrency
    volumes:
      - ./db-data/postgres/:/var/lib/postgresql/data/

  #  start Redis, and ensure that data is stored to a mounted volume
  redis:
    image: "redis:alpine"
    ports:
      - "6379:6379"
    restart: always
    volumes:
      - ./db-data/redis/:/data

  #  start mailhog
  mailhog:
    image: "mailhog/mailhog:latest"
    ports:
      - "1025:1025"
      - "8025:8025"
    restart: always
```

### Makefile

```makefile
BINARY_NAME=myapp
DSN="host=localhost port=5432 user=postgres password=password dbname=concurrency sslmode=disable timezone=UTC connect_timeout=5"
REDIS="127.0.0.1:6379"

## build: Build binary
build:
	@echo "Building..."
	env CGO_ENABLED=0  go build -ldflags="-s -w" -o ${BINARY_NAME} ./cmd/web
	@echo "Built!"

## run: builds and runs the application
run: build
	@echo "Starting..."
	@env DSN=${DSN} REDIS=${REDIS} ./${BINARY_NAME} &
	@echo "Started!"

## clean: runs go clean and deletes binaries
clean:
	@echo "Cleaning..."
	@go clean
	@rm ${BINARY_NAME}
	@echo "Cleaned!"

## start: an alias to run
start: run

## stop: stops the running application
stop:
	@echo "Stopping..."
	@-pkill -SIGTERM -f "./${BINARY_NAME}"
	@echo "Stopped!"

## restart: stops and starts the application
restart: stop start

## test: runs all tests
test:
	go test -v ./...
```

### Session

```go
func initSession() *scs.SessionManager {
	// set up session
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}
```

### Application config

```go
package main

import (
	"database/sql"
	"log"
	"sync"

	"github.com/alexedwards/scs/v2"
)

type Config struct {
	Session  *scs.SessionManager
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	Wait     *sync.WaitGroup
}
```

### Routes and handlers

```go
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	// create router
	mux := chi.NewRouter()

	// set up middleware
	mux.Use(middleware.Recoverer)

	// define application routes
	mux.Get("/", app.HomePage)

	return mux
}
```

### Render template

```go
package main

import "net/http"

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.gohtml", nil)
}
```

### Session middleware

```go
package main

import "net/http"

func (app *Config) SessionLoad(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}
```

### Graceful shutdown

```go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

func main() {
	// connect to the database
	db := initDB()

	// create sessions
	session := initSession()

	// create loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create channels

	// create waitgroup
	wg := sync.WaitGroup{}

	// set up the application config
	app := Config{
		Session:  session,
		DB:       db,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
		Wait:     &wg,
	}

	// set up mail

	// listen for signals
	go app.listenForShutdown()

	// listen for web connections
	app.serve()
}

func (app *Config) serve() {
	// start http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	app.InfoLog.Println("Starting web server...")
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

// initDB connects to Postgres and returns a pool of connections
func initDB() *sql.DB {
	conn := connectToDB()
	if conn == nil {
		log.Panic("can't connect to database")
	}
	return conn
}

// connectToDB tries to connect to postgres, and backs off until a connection
// is made, or we have not connected after 10 tries
func connectToDB() *sql.DB {
	counts := 0

	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not yet ready...")
		} else {
			log.Print("connected to database!")
			return connection
		}

		if counts > 10 {
			return nil
		}

		log.Print("Backing off for 1 second")
		time.Sleep(1 * time.Second)
		counts++

		continue
	}
}

// openDB opens a connection to Postgres, using a DSN read
// from the environment variable DSN
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// initSession sets up a session, using Redis for session store
func initSession() *scs.SessionManager {
	// set up session
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}

// initRedis returns a pool of connections to Redis using the
// environment variable REDIS
func initRedis() *redis.Pool {
	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}

	return redisPool
}

func (app *Config) listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.shutdown()
	os.Exit(0)
}

func (app *Config) shutdown() {
	// perform any cleanup tasks
	app.InfoLog.Println("would run cleanup tasks...")

	// block until waitgroup is empty
	app.Wait.Wait()

	app.InfoLog.Println("closing channels and shutting down application...")
}
```

### Populating database

```sql
--
-- Name: plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.plans (
                              id integer NOT NULL,
                              plan_name character varying(255),
                              plan_amount integer,
                              created_at timestamp without time zone,
                              updated_at timestamp without time zone
);


--
-- Name: plans_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.plans ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_plans (
                                   id integer NOT NULL,
                                   user_id integer,
                                   plan_id integer,
                                   created_at timestamp without time zone,
                                   updated_at timestamp without time zone
);


--
-- Name: user_plans_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.user_plans ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.user_plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


CREATE TABLE public.users (
                              id integer DEFAULT nextval('public.user_id_seq'::regclass) NOT NULL,
                              email character varying(255),
                              first_name character varying(255),
                              last_name character varying(255),
                              password character varying(60),
                              user_active integer DEFAULT 0,
                              is_admin integer default 0,
                              created_at timestamp without time zone,
                              updated_at timestamp without time zone
);


INSERT INTO "public"."users"("email","first_name","last_name","password","user_active", "is_admin", "created_at","updated_at")
VALUES
    (E'admin@example.com',E'Admin',E'User',E'$2a$12$1zGLuYDDNvATh4RA4avbKuheAMpb1svexSzrQm7up.bnpwQHs0jNe',1,1,E'2022-03-14 00:00:00',E'2022-03-14 00:00:00');

SELECT pg_catalog.setval('public.plans_id_seq', 1, false);


SELECT pg_catalog.setval('public.user_id_seq', 2, true);


SELECT pg_catalog.setval('public.user_plans_id_seq', 1, false);

INSERT INTO "public"."plans"("plan_name","plan_amount","created_at","updated_at")
VALUES
    (E'Bronze Plan',1000,E'2022-05-12 00:00:00',E'2022-05-12 00:00:00'),
    (E'Silver Plan',2000,E'2022-05-12 00:00:00',E'2022-05-12 00:00:00'),
    (E'Gold Plan',3000,E'2022-05-12 00:00:00',E'2022-05-12 00:00:00');


ALTER TABLE ONLY public.plans
    ADD CONSTRAINT plans_pkey PRIMARY KEY (id);


ALTER TABLE ONLY public.user_plans
    ADD CONSTRAINT user_plans_pkey PRIMARY KEY (id);


ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


ALTER TABLE ONLY public.user_plans
    ADD CONSTRAINT user_plans_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.plans(id) ON UPDATE RESTRICT ON DELETE CASCADE;


ALTER TABLE ONLY public.user_plans
    ADD CONSTRAINT user_plans_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE RESTRICT ON DELETE CASCADE;
```

### Data package

```go
package data

import (
	"database/sql"
	"time"
)

const dbTimeout = time.Second * 3

var db *sql.DB

// New is the function used to create an instance of the data package. It returns the type
// Model, which embeds all the types we want to be available to our application.
func New(dbPool *sql.DB) Models {
	db = dbPool

	return Models{
		User: User{},
		Plan: Plan{},
	}
}

// Models is the type for this package. Note that any model that is included as a member
// in this type is available to us throughout the application, anywhere that the
// app variable is used, provided that the model is also added in the New function.
type Models struct {
	User User
	Plan Plan
}
```

### Login/logout

```go
func (app *Config) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.RenewToken(r.Context())

	// parse form post
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	// get email and password from form post
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid credentials.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// check password
	validPassword, err := user.PasswordMatches(password)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid credentials.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !validPassword{
		app.Session.Put(r.Context(), "error", "Invalid credentials.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// okay, so log user in
	app.Session.Put(r.Context(), "userID", user.ID)
	app.Session.Put(r.Context(), "user", user)

	app.Session.Put(r.Context(), "flash", "Successful login!")

	// redirect the user
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {
	// clean up session
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
```

## Sending emails

### Mailer

[Simple email](https://github.com/xhit/go-simple-mail)

```go
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Wait        *sync.WaitGroup
	MailerChan  chan Message
	ErrorChan   chan error
	DoneChan    chan bool
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
	Template    string
}

// a function to listen for messages on the MailerChan

func (m *Mail) sendMail(msg Message, errorChan chan error) {
	if msg.Template == "" {
		msg.Template = "mail"
	}

	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data

	// build html mail
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		errorChan <- err
	}

	// build plain text mail
	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		errorChan <- err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		errorChan <- err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)

	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formattedMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		errorChan <- err
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.html.gohtml", msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.plain.gohtml", msg.Template)

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses: false,
		CssToAttributes: false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
```

### Send email sync

```go
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	// create router
	mux := chi.NewRouter()

	// set up middleware
	mux.Use(middleware.Recoverer)
	mux.Use(app.SessionLoad)

	// define application routes
	mux.Get("/", app.HomePage)

	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLoginPage)
	mux.Get("/logout", app.Logout)
	mux.Get("/register", app.RegisterPage)
	mux.Post("/register", app.PostRegisterPage)
	mux.Get("/activate-account", app.ActivateAccount)

	mux.Get("/test-email", func(w http.ResponseWriter, r *http.Request) {
		m := Mail{
			Domain: "localhost",
			Host: "localhost",
			Port: 1025,
			Encryption: "none",
			FromAddress: "info@mycompany.com",
			FromName: "info",
			ErrorChan: make(chan error),

		}

		msg := Message{
			To: "me@here.com",
			Subject: "Test email",
			Data: "Hello, world.",
		}

		m.sendMail(msg, make(chan error))
	})

	return mux
}
```
