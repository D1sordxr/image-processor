package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

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
