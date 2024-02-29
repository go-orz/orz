package orz

import "sync"

func NewTaskRunner[T any](max int) *Runner[T] {
	return &Runner[T]{
		ch: make(chan struct{}, max),
	}
}

type Runner[T any] struct {
	ch      chan struct{}
	wg      sync.WaitGroup
	results []RunnerResult[T]
	mux     sync.Mutex
}

type RunnerResult[T any] struct {
	Result T
	Error  error
}

func (r *Runner[T]) Execute(f func() (T, error)) {
	r.ch <- struct{}{}
	r.wg.Add(1)
	go func() {
		defer func() {
			<-r.ch
			r.wg.Done()
		}()
		t, err := f()
		r.addResult(t, err)
	}()
}

func (r *Runner[T]) Submit(f func() error) {
	r.ch <- struct{}{}
	r.wg.Add(1)
	go func() {
		defer func() {
			<-r.ch
			r.wg.Done()
		}()
		err := f()
		var t T
		r.addResult(t, err)
	}()
}

func (r *Runner[T]) addResult(t T, err error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.results = append(r.results, RunnerResult[T]{
		Result: t,
		Error:  err,
	})
}

func (r *Runner[T]) Wait() []RunnerResult[T] {
	r.wg.Wait()
	return r.results
}
