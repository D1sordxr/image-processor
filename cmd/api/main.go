package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	noCancel := context.WithoutCancel(ctx)
	go func() {
		<-noCancel.Done()
		fmt.Println("never done is done")
	}()

	<-ctx.Done()
	time.Sleep(3 * time.Second)
}
