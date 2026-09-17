package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/ekazakas/timewheel"
)

type (
	Executable func() error
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	startTime := time.Now()
	wheel := timewheel.NewTimingWheel[Executable](1*time.Second, 60, startTime)

	out := make(chan Executable)
	defer close(out)

	go func() {
		for o := range out {
			if err := o(); err != nil {
				log.Println(err)
			}
		}
	}()

	clock := timewheel.NewTickerClock(time.Second)
	defer clock.Stop()

	engine := timewheel.NewEngine(wheel, clock, out)

	go func() {
		if err := engine.Run(ctx); errors.Is(err, context.Canceled) {
			return
		} else if err != nil {
			panic(err)
		}
	}()

	for i := range 5 {
		if idx := engine.Schedule(time.Duration(i)*time.Second, func() error {
			log.Printf("Hello World %d!", i)

			return nil
		}); idx == timewheel.NullIndex {
			log.Println("Task was not scheduled")
		}
	}

	<-ctx.Done()

	log.Println("Engine exited")
}
