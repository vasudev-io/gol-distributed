package main

import (
	"fmt"
	"os"
	"testing"
	"uk.ac.bris.cs/gameoflife/gol"
)

const turnBench = 100

func BenchmarkGol(b *testing.B) {

	for threads := 1; threads <= 1; threads++ {
		os.Stdout = nil

		p := gol.Params{
			Turns:       turnBench,
			Threads:     threads,
			ImageWidth:  512,
			ImageHeight: 512,
		}

		benchName := fmt.Sprintf("%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)
		b.Run(benchName, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				events := make(chan gol.Event)
				go gol.Run(p, events, nil)

				for range events {

				}

			}

		})
	}
}
