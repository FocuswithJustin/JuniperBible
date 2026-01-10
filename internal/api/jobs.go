package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current state of a job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// Job represents an asynchronous conversion job.
type Job struct {
	ID          string                 `json:"id"`
	Status      JobStatus              `json:"status"`
	Progress    int                    `json:"progress"` // 0-100
	Result      *ConvertResult         `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	CompletedAt string                 `json:"completed_at,omitempty"`
	Request     ConvertRequest         `json:"request"`
	ctx         context.Context        `json:"-"`
	cancel      context.CancelFunc     `json:"-"`
	options     map[string]interface{} `json:"-"`
}

// JobStore manages conversion jobs in memory.
type JobStore struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

// NewJobStore creates a new job store.
func NewJobStore() *JobStore {
	return &JobStore{
		jobs: make(map[string]*Job),
	}
}

var globalJobStore = NewJobStore()

// Create creates a new job and returns its ID.
func (s *JobStore) Create(req ConvertRequest) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now().UTC().Format(time.RFC3339)

	job := &Job{
		ID:        uuid.New().String(),
		Status:    JobStatusPending,
		Progress:  0,
		CreatedAt: now,
		UpdatedAt: now,
		Request:   req,
		ctx:       ctx,
		cancel:    cancel,
		options:   req.Options,
	}

	s.jobs[job.ID] = job
	return job
}

// Get retrieves a job by ID.
func (s *JobStore) Get(id string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	return job, exists
}

// Update updates a job's status and progress.
func (s *JobStore) Update(id string, status JobStatus, progress int, result *ConvertResult, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}

	job.Status = status
	job.Progress = progress
	job.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if result != nil {
		job.Result = result
	}

	if errMsg != "" {
		job.Error = errMsg
	}

	if status == JobStatusCompleted || status == JobStatusFailed || status == JobStatusCancelled {
		job.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	}

	return nil
}

// Delete removes a job from the store.
func (s *JobStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}

	// Cancel if still running
	if job.Status == JobStatusRunning || job.Status == JobStatusPending {
		if job.cancel != nil {
			job.cancel()
		}
	}

	delete(s.jobs, id)
	return nil
}

// List returns all jobs.
func (s *JobStore) List() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// Cancel cancels a running job.
func (s *JobStore) Cancel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}

	if job.Status != JobStatusPending && job.Status != JobStatusRunning {
		return fmt.Errorf("job cannot be cancelled (status: %s)", job.Status)
	}

	if job.cancel != nil {
		job.cancel()
	}

	job.Status = JobStatusCancelled
	job.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	job.CompletedAt = time.Now().UTC().Format(time.RFC3339)

	return nil
}

// runJob executes a conversion job in a goroutine.
func runJob(job *Job) {
	go func() {
		// Update status to running
		globalJobStore.Update(job.ID, JobStatusRunning, 10, nil, "")

		// Simulate conversion work with progress updates
		// In real implementation, this would call actual conversion logic
		for i := 10; i <= 90; i += 20 {
			select {
			case <-job.ctx.Done():
				// Job was cancelled
				globalJobStore.Update(job.ID, JobStatusCancelled, i, nil, "Job cancelled by user")
				return
			default:
				time.Sleep(500 * time.Millisecond)
				globalJobStore.Update(job.ID, JobStatusRunning, i, nil, "")
			}
		}

		// Check for cancellation before completing
		select {
		case <-job.ctx.Done():
			globalJobStore.Update(job.ID, JobStatusCancelled, 90, nil, "Job cancelled by user")
			return
		default:
		}

		// For now, return NOT_IMPLEMENTED as the actual conversion is not yet implemented
		// In a real implementation, this would perform the actual conversion and return the result

		// Mark as failed since conversion is not yet implemented
		globalJobStore.Update(job.ID, JobStatusFailed, 100, nil,
			fmt.Sprintf("Conversion from %s to %s not yet implemented via API. Use the CLI.",
				job.Request.Source, job.Request.TargetFormat))
	}()
}

// handleJobs handles POST /jobs - Create new conversion job.
func handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed")
		return
	}

	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	if req.Source == "" || req.TargetFormat == "" {
		respondError(w, http.StatusBadRequest, "MISSING_PARAMS", "source and target_format are required")
		return
	}

	// Create job
	job := globalJobStore.Create(req)

	// Start job in background
	runJob(job)

	// Return job info
	respond(w, http.StatusCreated, job)
}

// handleJobByID handles GET /jobs/{id} - Get job status and DELETE /jobs/{id} - Cancel job.
func handleJobByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if id == "" {
		respondError(w, http.StatusBadRequest, "MISSING_ID", "Job ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		getJobHandler(w, r, id)
	case http.MethodDelete:
		cancelJobHandler(w, r, id)
	default:
		respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET and DELETE are allowed")
	}
}

// getJobHandler handles GET /jobs/{id}.
func getJobHandler(w http.ResponseWriter, r *http.Request, id string) {
	job, exists := globalJobStore.Get(id)
	if !exists {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "Job not found")
		return
	}

	respond(w, http.StatusOK, job)
}

// cancelJobHandler handles DELETE /jobs/{id}.
func cancelJobHandler(w http.ResponseWriter, r *http.Request, id string) {
	if err := globalJobStore.Cancel(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, "CANCEL_FAILED", err.Error())
		return
	}

	respond(w, http.StatusOK, map[string]string{"message": "Job cancelled"})
}
