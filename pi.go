/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE', which is part of this source code package.
 */
package main

import (
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"strconv"
)

var (
	precision uint
	minusOne  = big.NewFloat(-1)
	one       = big.NewFloat(1)
	two       = big.NewFloat(2)
	three     = big.NewFloat(3)
	four      = big.NewFloat(4)
)

func main() {

	flag.UintVar(&precision, "p", 64, "Precision of the resulting Number")
	flag.Parse()
	rounds, err := strconv.ParseInt(flag.Arg(0), 10, 64)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Usage: pi <rounds>")
		os.Exit(1)
	}

	pi := calcpi(rounds)

	fmt.Println(pi.Text('f', int(precision)))
	fmt.Printf("prec = %d, acc = %s\n", pi.Prec(), pi.Acc())
}

func calcpi(n int64) *big.Float {

	terms := make(chan *big.Float, runtime.NumCPU())
	input := make(chan int64, runtime.NumCPU())

	for i := 0; i < runtime.NumCPU()*2; i++ {
		go termCalculation(terms, input)
	}

	go fillInput(input, n)

	result := big.NewFloat(float64(0.0)).SetPrec(precision)

	for i := int64(0); i < n; i++ {
		term := <-terms
		result.Add(result, term)
	}

	return result
}

func fillInput(input chan<- int64, n int64) {
	for i := int64(0); i < n; i++ {
		input <- i
	}
	close(input)
}

func termCalculation(terms chan<- *big.Float, input <-chan int64) {

	for {
		select {
		case k, ok := <-input:

			if !ok {
				return
			}

			bbpTerm(terms, k)
		}
	}
}

func bbpTerm(terms chan<- *big.Float, k int64) {

	fourTimesK := new(big.Float).SetInt64(k)
	fourTimesK.Mul(fourTimesK, four)

	// (-1)^k / 4^k
	multiplicand := new(big.Float).SetPrec(precision).Quo(new(big.Float).SetFloat64(math.Pow(-1, float64(k))), new(big.Float).SetFloat64(math.Pow(4, float64(k))))

	// 2 / (4*k+1)
	firstSubTerm := new(big.Float).SetPrec(precision).Quo(two, new(big.Float).Add(fourTimesK, one))
	// 2 / (4*k+2)
	secondSubTerm := new(big.Float).SetPrec(precision).Quo(two, new(big.Float).Add(fourTimesK, two))
	// 1 / (4*k+3)
	thirdSubTerm := new(big.Float).SetPrec(precision).Quo(one, new(big.Float).Add(fourTimesK, three))
	// (2 / (4*k+1) + 2 / (4*k+2) + 1 / (4*k+3))
	parenthesis := new(big.Float).SetPrec(precision).Add(firstSubTerm, secondSubTerm)
	parenthesis.Add(parenthesis, thirdSubTerm)

	// fmt.Printf("k = %d: multiplicand = %s; firstSubTerm = %s; secondSubTerm = %s; thirdSubTerm = %s; parenthesis = %s\n", k, multiplicand.Text('f', 5), firstSubTerm.Text('f', 5), secondSubTerm.Text('f', 5), thirdSubTerm.Text('f', 5), parenthesis.Text('f', 5))

	// ((-1)^k / 4^k) * (2 / (4*k+1) + 2 / (4*k+2) + 1 / (4*k+3))
	terms <- new(big.Float).SetPrec(precision).Mul(multiplicand, parenthesis)
}
