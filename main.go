package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nitin1chandani/ticketmaster/app"
	"github.com/nitin1chandani/ticketmaster/internal/worker"
)

func main() {
	container, err := app.NewContainer()
	if err != nil {
		log.Fatal(err)
	}
	defer container.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	expiryWorker := worker.NewReservationExpiryWorker(container.DB, 15*time.Second, 5*time.Second, slog.Default())
	go expiryWorker.Start(ctx)

	go func() {
		<-ctx.Done()
		_ = container.App.Shutdown()
	}()
	log.Fatal(container.App.Listen(":" + container.Config.Port))
}
