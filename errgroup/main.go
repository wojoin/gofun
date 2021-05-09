package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"golang.org/x/sync/errgroup"
	"syscall"
)


type Result string
type Service func(ctx context.Context, serviceName string) (Result, error)

type srv func()

var (
	httpServer   = fakeService("http server", httpSrv)
	linuxSig = fakeService("signal", linuxSignal)
)


func fakeService(kind string, s srv) Service {
	return func(_ context.Context, serviceName string) (Result, error) {
		var res Result
		if s != nil {
			s()
			res = Result(fmt.Sprintf("%s result for %q", kind, serviceName))
		}
		return  res, nil
	}
}


func httpSrv() {
	m := http.NewServeMux()
	s := http.Server{Addr: ":8080", Handler: m}
	m.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ShutDown"))
		go func() {
			if err := s.Shutdown(context.Background()); err != nil {
				log.Fatal(err)
			}
		}()
	})
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Printf("Finished")
}

func linuxSignal() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		done <- true
	}()
	fmt.Println("awaiting signal")
	<-done
	fmt.Println("exiting")
}


func main() {
	Google := func(ctx context.Context, serviceName string) ([]Result, error) {
		g, ctx := errgroup.WithContext(ctx)

		services := []Service{httpServer, linuxSig}
		results := make([]Result, len(services))
		for i, service := range services {
			i, service := i, service
			g.Go(func() error {
				result, err := service(ctx, serviceName)
				if err == nil {
					results[i] = result
				}
				return err
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}
		return results, nil
	}

	results, err := Google(context.Background(), "golang")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, result := range results {
		fmt.Println(result)
	}
}
