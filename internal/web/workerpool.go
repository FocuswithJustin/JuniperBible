package web

import "sync"

// WorkerPool provides a reusable worker pool pattern for parallel processing.
// It manages job distribution across multiple workers and collects results.
type WorkerPool[Job any, Result any] struct {
	numWorkers int
	jobs       chan Job
	results    chan Result
	wg         sync.WaitGroup
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
// If numWorkers is 0 or negative, it defaults to maxWorkers.
// If numJobs is less than numWorkers, the pool is sized to match numJobs.
func NewWorkerPool[Job any, Result any](numWorkers, numJobs int) *WorkerPool[Job, Result] {
	if numWorkers <= 0 {
		numWorkers = maxWorkers
	}
	if numJobs > 0 {
		numWorkers = min(numWorkers, numJobs)
	}

	return &WorkerPool[Job, Result]{
		numWorkers: numWorkers,
		jobs:       make(chan Job, numJobs),
		results:    make(chan Result, numJobs),
	}
}

// Start begins the worker pool with the provided worker function.
// The workerFn is called for each job and should return a result.
func (p *WorkerPool[Job, Result]) Start(workerFn func(Job) Result) {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for job := range p.jobs {
				result := workerFn(job)
				p.results <- result
			}
		}()
	}
}

// Submit adds a job to the worker pool's job queue.
func (p *WorkerPool[Job, Result]) Submit(job Job) {
	p.jobs <- job
}

// Close closes the job channel and waits for all workers to complete.
// After calling Close, the results channel will be closed automatically.
func (p *WorkerPool[Job, Result]) Close() {
	close(p.jobs)
	go func() {
		p.wg.Wait()
		close(p.results)
	}()
}

// Results returns the results channel for collecting worker outputs.
func (p *WorkerPool[Job, Result]) Results() <-chan Result {
	return p.results
}
