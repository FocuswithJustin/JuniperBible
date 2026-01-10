package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandleJobsMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	w := httptest.NewRecorder()

	handleJobs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if apiResp.Success {
		t.Error("expected success to be false")
	}

	if apiResp.Error == nil || apiResp.Error.Code != "METHOD_NOT_ALLOWED" {
		t.Error("expected METHOD_NOT_ALLOWED error")
	}
}

func TestHandleJobsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleJobs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if apiResp.Error == nil || apiResp.Error.Code != "INVALID_JSON" {
		t.Error("expected INVALID_JSON error")
	}
}

func TestHandleJobsMissingParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleJobs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if apiResp.Error == nil || apiResp.Error.Code != "MISSING_PARAMS" {
		t.Error("expected MISSING_PARAMS error")
	}
}

func TestHandleJobsMissingSource(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"target_format":"usfm"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleJobs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleJobsMissingTargetFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"source":"test.osis"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleJobs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleJobsCreateSuccess(t *testing.T) {
	// Clear global job store
	globalJobStore = NewJobStore()

	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"source":"test.osis","target_format":"usfm"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleJobs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Error("expected success to be true")
	}

	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}

	if data["id"] == nil || data["id"] == "" {
		t.Error("expected job ID to be set")
	}

	if data["status"] == nil {
		t.Error("expected job status to be set")
	}

	if data["created_at"] == nil {
		t.Error("expected job created_at to be set")
	}

	if data["progress"] == nil {
		t.Error("expected job progress to be set")
	}
}

func TestHandleJobByIDMissingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/jobs/", nil)
	w := httptest.NewRecorder()

	handleJobByID(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if apiResp.Error == nil || apiResp.Error.Code != "MISSING_ID" {
		t.Error("expected MISSING_ID error")
	}
}

func TestHandleJobByIDMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/jobs/test-id", nil)
	w := httptest.NewRecorder()

	handleJobByID(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", resp.StatusCode)
	}
}

func TestGetJobHandlerNotFound(t *testing.T) {
	// Clear global job store
	globalJobStore = NewJobStore()

	req := httptest.NewRequest(http.MethodGet, "/jobs/nonexistent-id", nil)
	w := httptest.NewRecorder()

	handleJobByID(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if apiResp.Error == nil || apiResp.Error.Code != "NOT_FOUND" {
		t.Error("expected NOT_FOUND error")
	}
}

func TestGetJobHandlerSuccess(t *testing.T) {
	// Clear global job store
	globalJobStore = NewJobStore()

	// Create a job
	job := globalJobStore.Create(ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	})

	req := httptest.NewRequest(http.MethodGet, "/jobs/"+job.ID, nil)
	w := httptest.NewRecorder()

	handleJobByID(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Error("expected success to be true")
	}

	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}

	if data["id"] != job.ID {
		t.Errorf("expected job ID %s, got %v", job.ID, data["id"])
	}
}

func TestCancelJobHandlerNotFound(t *testing.T) {
	// Clear global job store
	globalJobStore = NewJobStore()

	req := httptest.NewRequest(http.MethodDelete, "/jobs/nonexistent-id", nil)
	w := httptest.NewRecorder()

	handleJobByID(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if apiResp.Error == nil || apiResp.Error.Code != "NOT_FOUND" {
		t.Error("expected NOT_FOUND error")
	}
}

func TestCancelJobHandlerSuccess(t *testing.T) {
	// Clear global job store
	globalJobStore = NewJobStore()

	// Create a job
	job := globalJobStore.Create(ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	})

	// Start the job
	runJob(job)
	time.Sleep(100 * time.Millisecond) // Give it time to start

	req := httptest.NewRequest(http.MethodDelete, "/jobs/"+job.ID, nil)
	w := httptest.NewRecorder()

	handleJobByID(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Error("expected success to be true")
	}

	// Verify job was cancelled
	cancelledJob, _ := globalJobStore.Get(job.ID)
	if cancelledJob.Status != JobStatusCancelled {
		t.Errorf("expected job status to be cancelled, got %s", cancelledJob.Status)
	}
}

func TestCancelJobHandlerAlreadyCompleted(t *testing.T) {
	// Clear global job store
	globalJobStore = NewJobStore()

	// Create a completed job
	job := globalJobStore.Create(ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	})
	globalJobStore.Update(job.ID, JobStatusCompleted, 100, &ConvertResult{
		OutputPath: "/test/output.usfm",
		LossClass:  "L0",
		Duration:   "1s",
	}, "")

	req := httptest.NewRequest(http.MethodDelete, "/jobs/"+job.ID, nil)
	w := httptest.NewRecorder()

	handleJobByID(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if apiResp.Error == nil || apiResp.Error.Code != "CANCEL_FAILED" {
		t.Error("expected CANCEL_FAILED error")
	}
}

func TestJobStoreCreate(t *testing.T) {
	store := NewJobStore()

	req := ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
		Options:      map[string]interface{}{"key": "value"},
	}

	job := store.Create(req)

	if job.ID == "" {
		t.Error("expected job ID to be set")
	}

	if job.Status != JobStatusPending {
		t.Errorf("expected job status to be pending, got %s", job.Status)
	}

	if job.Progress != 0 {
		t.Errorf("expected job progress to be 0, got %d", job.Progress)
	}

	if job.CreatedAt == "" {
		t.Error("expected job created_at to be set")
	}

	if job.Request.Source != req.Source {
		t.Errorf("expected source %s, got %s", req.Source, job.Request.Source)
	}

	if job.Request.TargetFormat != req.TargetFormat {
		t.Errorf("expected target_format %s, got %s", req.TargetFormat, job.Request.TargetFormat)
	}
}

func TestJobStoreGet(t *testing.T) {
	store := NewJobStore()

	req := ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	}

	job := store.Create(req)

	// Get existing job
	retrieved, exists := store.Get(job.ID)
	if !exists {
		t.Error("expected job to exist")
	}

	if retrieved.ID != job.ID {
		t.Errorf("expected job ID %s, got %s", job.ID, retrieved.ID)
	}

	// Get non-existent job
	_, exists = store.Get("nonexistent-id")
	if exists {
		t.Error("expected job to not exist")
	}
}

func TestJobStoreUpdate(t *testing.T) {
	store := NewJobStore()

	req := ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	}

	job := store.Create(req)

	// Update job
	result := &ConvertResult{
		OutputPath: "/test/output.usfm",
		LossClass:  "L0",
		Duration:   "1s",
	}

	err := store.Update(job.ID, JobStatusCompleted, 100, result, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify update
	updated, _ := store.Get(job.ID)
	if updated.Status != JobStatusCompleted {
		t.Errorf("expected status completed, got %s", updated.Status)
	}

	if updated.Progress != 100 {
		t.Errorf("expected progress 100, got %d", updated.Progress)
	}

	if updated.Result == nil {
		t.Fatal("expected result to be set")
	}

	if updated.Result.OutputPath != result.OutputPath {
		t.Errorf("expected output path %s, got %s", result.OutputPath, updated.Result.OutputPath)
	}

	if updated.CompletedAt == "" {
		t.Error("expected completed_at to be set")
	}

	// Update non-existent job
	err = store.Update("nonexistent-id", JobStatusCompleted, 100, nil, "")
	if err == nil {
		t.Error("expected error when updating non-existent job")
	}
}

func TestJobStoreUpdateWithError(t *testing.T) {
	store := NewJobStore()

	req := ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	}

	job := store.Create(req)

	// Update job with error
	errMsg := "conversion failed"
	err := store.Update(job.ID, JobStatusFailed, 50, nil, errMsg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify update
	updated, _ := store.Get(job.ID)
	if updated.Status != JobStatusFailed {
		t.Errorf("expected status failed, got %s", updated.Status)
	}

	if updated.Error != errMsg {
		t.Errorf("expected error %s, got %s", errMsg, updated.Error)
	}

	if updated.CompletedAt == "" {
		t.Error("expected completed_at to be set for failed job")
	}
}

func TestJobStoreDelete(t *testing.T) {
	store := NewJobStore()

	req := ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	}

	job := store.Create(req)

	// Delete existing job
	err := store.Delete(job.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify deletion
	_, exists := store.Get(job.ID)
	if exists {
		t.Error("expected job to be deleted")
	}

	// Delete non-existent job
	err = store.Delete("nonexistent-id")
	if err == nil {
		t.Error("expected error when deleting non-existent job")
	}
}

func TestJobStoreList(t *testing.T) {
	store := NewJobStore()

	// Empty store
	jobs := store.List()
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(jobs))
	}

	// Add jobs
	store.Create(ConvertRequest{Source: "test1.osis", TargetFormat: "usfm"})
	store.Create(ConvertRequest{Source: "test2.osis", TargetFormat: "json"})

	jobs = store.List()
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestJobStoreCancel(t *testing.T) {
	store := NewJobStore()

	req := ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	}

	job := store.Create(req)

	// Start job
	store.Update(job.ID, JobStatusRunning, 50, nil, "")

	// Cancel job
	err := store.Cancel(job.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify cancellation
	cancelled, _ := store.Get(job.ID)
	if cancelled.Status != JobStatusCancelled {
		t.Errorf("expected status cancelled, got %s", cancelled.Status)
	}

	if cancelled.CompletedAt == "" {
		t.Error("expected completed_at to be set")
	}

	// Cancel non-existent job
	err = store.Cancel("nonexistent-id")
	if err == nil {
		t.Error("expected error when cancelling non-existent job")
	}
}

func TestJobStoreCancelCompleted(t *testing.T) {
	store := NewJobStore()

	req := ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	}

	job := store.Create(req)

	// Complete job
	store.Update(job.ID, JobStatusCompleted, 100, &ConvertResult{
		OutputPath: "/test/output.usfm",
		LossClass:  "L0",
		Duration:   "1s",
	}, "")

	// Try to cancel completed job
	err := store.Cancel(job.ID)
	if err == nil {
		t.Error("expected error when cancelling completed job")
	}
}

func TestRunJobProgress(t *testing.T) {
	// Clear global job store
	globalJobStore = NewJobStore()

	// Create a job
	job := globalJobStore.Create(ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	})

	// Start job
	runJob(job)

	// Wait for job to start
	time.Sleep(100 * time.Millisecond)

	// Check job is running
	running, _ := globalJobStore.Get(job.ID)
	if running.Status != JobStatusRunning && running.Status != JobStatusFailed {
		t.Errorf("expected status running or failed, got %s", running.Status)
	}

	if running.Progress <= 0 {
		t.Error("expected progress to be greater than 0")
	}

	// Wait for job to complete
	time.Sleep(3 * time.Second)

	// Check job is completed or failed
	completed, _ := globalJobStore.Get(job.ID)
	if completed.Status != JobStatusCompleted && completed.Status != JobStatusFailed {
		t.Errorf("expected status completed or failed, got %s", completed.Status)
	}

	if completed.Progress != 100 {
		t.Errorf("expected progress 100, got %d", completed.Progress)
	}

	if completed.CompletedAt == "" {
		t.Error("expected completed_at to be set")
	}
}

func TestRunJobCancellation(t *testing.T) {
	// Clear global job store
	globalJobStore = NewJobStore()

	// Create a job
	job := globalJobStore.Create(ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	})

	// Start job
	runJob(job)

	// Wait for job to start
	time.Sleep(200 * time.Millisecond)

	// Cancel job
	err := globalJobStore.Cancel(job.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Wait a bit for cancellation to take effect
	time.Sleep(200 * time.Millisecond)

	// Verify job was cancelled
	cancelled, _ := globalJobStore.Get(job.ID)
	if cancelled.Status != JobStatusCancelled {
		t.Errorf("expected status cancelled, got %s", cancelled.Status)
	}
}

func TestJobStoreDeleteRunningJob(t *testing.T) {
	// Clear global job store
	globalJobStore = NewJobStore()

	// Create a job
	job := globalJobStore.Create(ConvertRequest{
		Source:       "test.osis",
		TargetFormat: "usfm",
	})

	// Start job
	runJob(job)
	time.Sleep(100 * time.Millisecond)

	// Delete running job (should cancel it)
	err := globalJobStore.Delete(job.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify deletion
	_, exists := globalJobStore.Get(job.ID)
	if exists {
		t.Error("expected job to be deleted")
	}
}
