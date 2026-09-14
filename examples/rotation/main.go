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

	wheel := timewheel.NewTimingWheel[Executable](1*time.Second, time.Now(), 60)

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

	initialTask := timewheel.NewTask[Executable](time.Now().Add(3*time.Second), func() error {
		log.Println("This shouldn't print if rescheduled successfully!")

		return nil
	})

	if node := wheel.Add(initialTask); node != nil {
		log.Println("Scheduled task")

		time.Sleep(1 * time.Second)

		if wheel.Remove(node) {
			log.Println("Removed task")
		}
	}

	updatedTask := timewheel.NewTask[Executable](time.Now().Add(7*time.Second), func() error {
		log.Println("Success! The updated task was executed at its new prolonged time.")

		return nil
	})

	if newNode := wheel.Add(updatedTask); newNode != nil {
		log.Println("Rescheduled task")
	}

	<-ctx.Done()

	log.Println("Engine exited.")
}
