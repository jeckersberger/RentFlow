package repositories

import (
	"context"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// NoopEquipmentLocker is a placeholder implementation that logs lock/unlock events
// In production, this would call the equipment-service via gRPC or HTTP
type NoopEquipmentLocker struct {
	logger logger.Logger
}

func NewNoopEquipmentLocker(log logger.Logger) *NoopEquipmentLocker {
	return &NoopEquipmentLocker{logger: log}
}

func (n *NoopEquipmentLocker) LockForMaintenance(ctx context.Context, tenantID, equipmentID, taskID string) error {
	n.logger.Info("Equipment locked for maintenance", "equipment_id", equipmentID, "task_id", taskID)
	// TODO: Call equipment-service to set status to "maintenance"
	return nil
}

func (n *NoopEquipmentLocker) UnlockFromMaintenance(ctx context.Context, tenantID, equipmentID, taskID string) error {
	n.logger.Info("Equipment unlocked from maintenance", "equipment_id", equipmentID, "task_id", taskID)
	// TODO: Call equipment-service to restore status to "available"
	return nil
}

func (n *NoopEquipmentLocker) IsLockedForMaintenance(ctx context.Context, tenantID, equipmentID string) (bool, error) {
	// TODO: Call equipment-service to check maintenance status
	return false, nil
}
