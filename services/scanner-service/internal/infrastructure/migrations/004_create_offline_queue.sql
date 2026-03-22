CREATE TABLE IF NOT EXISTS offline_queue (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    payload TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    synced_at TIMESTAMP,
    sync_status VARCHAR(50) NOT NULL,

    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_offline_queue_tenant_id ON offline_queue(tenant_id);
CREATE INDEX idx_offline_queue_device_id ON offline_queue(device_id);
CREATE INDEX idx_offline_queue_sync_status ON offline_queue(sync_status);
CREATE INDEX idx_offline_queue_pending ON offline_queue(tenant_id, sync_status) WHERE sync_status IN ('pending', 'failed');
CREATE INDEX idx_offline_queue_created_at ON offline_queue(created_at DESC);
