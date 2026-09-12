package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/ports"
)

type retryScanRepo struct {
	events  map[string]*domain.ScanEvent
	creates int
}

func newRetryScanRepo() *retryScanRepo {
	return &retryScanRepo{events: make(map[string]*domain.ScanEvent)}
}

func (r *retryScanRepo) Create(context.Context, *domain.ScanEvent) error { return nil }

func (r *retryScanRepo) CreateIfAbsent(_ context.Context, event *domain.ScanEvent) (bool, error) {
	if _, exists := r.events[event.ID]; exists {
		return false, nil
	}
	copy := *event
	r.events[event.ID] = &copy
	r.creates++
	return true, nil
}

func (r *retryScanRepo) GetByID(_ context.Context, tenantID, id string) (*domain.ScanEvent, error) {
	event, ok := r.events[id]
	if !ok || event.TenantID != tenantID {
		return nil, errors.New("not found")
	}
	copy := *event
	return &copy, nil
}

func (r *retryScanRepo) List(context.Context, string, *ports.ScanListQuery) (*ports.ScanListResult, error) {
	return &ports.ScanListResult{}, nil
}
func (r *retryScanRepo) Update(context.Context, *domain.ScanEvent) error { return nil }
func (r *retryScanRepo) GetByBarcode(context.Context, string, string) (*domain.ScanEvent, error) {
	return nil, errors.New("not found")
}

type retryQueueRepo struct {
	stored          domain.OfflineQueueItem
	failNextUpdate bool
}

func (r *retryQueueRepo) Create(_ context.Context, item *domain.OfflineQueueItem) error {
	r.stored = *item
	return nil
}
func (r *retryQueueRepo) GetByID(_ context.Context, id string) (*domain.OfflineQueueItem, error) {
	if r.stored.ID != id {
		return nil, errors.New("not found")
	}
	copy := r.stored
	return &copy, nil
}
func (r *retryQueueRepo) GetPending(_ context.Context, tenantID string, _ int) ([]*domain.OfflineQueueItem, error) {
	if r.stored.TenantID != tenantID || r.stored.SyncStatus != "pending" {
		return nil, nil
	}
	copy := r.stored
	return []*domain.OfflineQueueItem{&copy}, nil
}
func (r *retryQueueRepo) Update(_ context.Context, item *domain.OfflineQueueItem) error {
	if r.failNextUpdate {
		r.failNextUpdate = false
		return errors.New("simulated acknowledgement failure")
	}
	r.stored = *item
	return nil
}
func (r *retryQueueRepo) DeleteByID(context.Context, string) error { return nil }
func (r *retryQueueRepo) GetCount(context.Context, string, string) (int, error) { return 0, nil }

func TestSyncOfflineQueueRetryDoesNotDuplicateScan(t *testing.T) {
	ctx := context.Background()
	scanRepo := newRetryScanRepo()

	cmd := ProcessScanCommand{
		Barcode:    "123456789",
		ScanType:   domain.ScanCheckOut,
		UserID:     "user-1",
		DeviceID:   "scanner-1",
		DeviceType: domain.DeviceHandheld,
	}
	payload, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("marshal scan command: %v", err)
	}

	queueRepo := &retryQueueRepo{
		stored: domain.OfflineQueueItem{
			ID:         "queue_retry_1",
			TenantID:   "tenant-1",
			DeviceID:   "scanner-1",
			Payload:    string(payload),
			SyncStatus: "pending",
		},
		failNextUpdate: true,
	}

	svc := NewSessionService(nil, scanRepo, queueRepo, nil, nil, logger.New("error", "scanner-test"))

	first, err := svc.SyncOfflineQueue(ctx, "tenant-1", 100)
	if err != nil {
		t.Fatalf("first sync returned error: %v", err)
	}
	if first.SyncedItems != 0 || first.FailedItems != 1 {
		t.Fatalf("first sync = synced %d failed %d, want 0/1", first.SyncedItems, first.FailedItems)
	}
	if scanRepo.creates != 1 {
		t.Fatalf("scan creates after first attempt = %d, want 1", scanRepo.creates)
	}
	if queueRepo.stored.SyncStatus != "pending" {
		t.Fatalf("queue status after failed acknowledgement = %q, want pending", queueRepo.stored.SyncStatus)
	}

	second, err := svc.SyncOfflineQueue(ctx, "tenant-1", 100)
	if err != nil {
		t.Fatalf("second sync returned error: %v", err)
	}
	if second.SyncedItems != 1 || second.FailedItems != 0 {
		t.Fatalf("second sync = synced %d failed %d, want 1/0", second.SyncedItems, second.FailedItems)
	}
	if scanRepo.creates != 1 {
		t.Fatalf("scan creates after retry = %d, want exactly 1", scanRepo.creates)
	}
	if got := len(scanRepo.events); got != 1 {
		t.Fatalf("persisted scan events = %d, want exactly 1", got)
	}
	if _, ok := scanRepo.events[offlineScanID("queue_retry_1")]; !ok {
		t.Fatalf("expected deterministic scan ID %q", offlineScanID("queue_retry_1"))
	}
	if queueRepo.stored.SyncStatus != "synced" {
		t.Fatalf("queue status after retry = %q, want synced", queueRepo.stored.SyncStatus)
	}
}
