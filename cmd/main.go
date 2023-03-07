package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"myhttp"
	"os"

	"golang.org/x/sync/errgroup"
)

func main() {
	fParallel := flag.Int("parallel", 10, "the number of parallel requests,")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("please enter url(s) for md5 conversion")
		os.Exit(1)
	}

	if *fParallel <= 0 {
		fmt.Println("please enter correct number of parallel requests")
		os.Exit(1)
	}

	s := myhttp.New(*fParallel)
	defer s.Close()

	// use error group to process input and output and process goroutines
	g, ctx := errgroup.WithContext(context.Background())

	// at this point run doesn't return any errors
	g.Go(func() error {
		s.Run(ctx)
		return nil
	})

	urlArgs := flag.Args()
	g.Go(func() error {
		for _, arg := range urlArgs {
			s.Send(arg)
		}
		return nil
	})

	g.Go(func() error {
		left := len(urlArgs)
		for {
			if left == 0 {
				return errors.New("no more urls left")
			}

			select {
			case <-ctx.Done():
				fmt.Println("program cancelled")
				return nil

			case resp := <-s.Recv():
				left--
				if resp.Err == nil {
					fmt.Println(resp.OriginWithScheme, resp.Encoded)
					continue
				}
				fmt.Printf("could not encode %s: %s\n", resp.OriginWithScheme, resp.Err)
			}
		}
	})

	_ = g.Wait()
}
