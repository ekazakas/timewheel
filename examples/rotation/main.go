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

	wheel := timewheel.NewTimingWheel[Executable](1*time.Second, 60, time.Now())

	out := make(chan Executable)
	defer close(out)

	go func() {
		for executionBlock := range out {
			if err := executionBlock(); err != nil {
				log.Println("Execution error:", err)
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

	if idx := engine.Schedule(time.Duration(3)*time.Second, func() error {
		log.Println("This shouldn't print if rescheduled successfully!")

		return nil
	}); idx == timewheel.NullIndex {
		log.Println("Initial task was not scheduled")
	} else {
		time.Sleep(1 * time.Second)

		if engine.Remove(idx) {
			log.Println("Removed task")
		}
	}

	if idx := engine.Schedule(3*time.Second, func() error {
		log.Println("Success! The updated task was executed at its new prolonged time.")

		return nil
	}); idx == timewheel.NullIndex {
		log.Println("Updated task was not scheduled")
	}

	<-ctx.Done()

	log.Println("Engine exited.")
}
