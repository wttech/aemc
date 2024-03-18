package lox

import (
	"context"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"time"
)

func Map[I any, R any](parallel bool, iterable []I, callback func(instance I) (R, error)) ([]R, error) {
	if parallel {
		return ParallelMap(iterable, callback)
	}
	return SerialMap(iterable, callback)
}

func ParallelMap[I any, R any](iterable []I, callback func(iteratee I) (R, error)) ([]R, error) {
	g, _ := errgroup.WithContext(context.Background())
	results := make([]R, len(iterable))
	for i, iteratee := range iterable {
		i, iteratee := i, iteratee
		g.Go(func() error {
			result, err := callback(iteratee)
			if err != nil {
				return err
			}
			results[i] = result
			return nil
		})
	}
	err := g.Wait()
	return results, err
}

func SerialMap[I any, R any](iterable []I, callback func(iteratee I) (R, error)) ([]R, error) {
	results := make([]R, len(iterable))
	for i, iteratee := range iterable {
		result, err := callback(iteratee)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func Random[I any](iterable []I) I {
	if len(iterable) == 0 {
		panic("cannot get random value from empty slice")
	}
	return iterable[random.Intn(len(iterable))]
}
