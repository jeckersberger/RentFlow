-- Federation Service Schema

CREATE TABLE IF NOT EXISTS federation_peers (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    public_key TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    last_seen TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_federation_peers_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT uk_federation_peers_url UNIQUE(tenant_id, url)
);

CREATE TABLE IF NOT EXISTS federation_equipment_shares (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    equipment_id VARCHAR(255) NOT NULL,
    peer_id VARCHAR(255) NOT NULL,
    availability VARCHAR(100) NOT NULL,
    price_per_day DECIMAL(10, 2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_federation_equipment_shares_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_federation_equipment_shares_peer FOREIGN KEY (peer_id) REFERENCES federation_peers(id)
);

CREATE TABLE IF NOT EXISTS federation_share_requests (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    from_peer_id VARCHAR(255) NOT NULL,
    to_peer_id VARCHAR(255) NOT NULL,
    equipment_id VARCHAR(255) NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_federation_share_requests_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_federation_share_requests_from FOREIGN KEY (from_peer_id) REFERENCES federation_peers(id),
    CONSTRAINT fk_federation_share_requests_to FOREIGN KEY (to_peer_id) REFERENCES federation_peers(id)
);

CREATE INDEX idx_federation_peers_tenant ON federation_peers(tenant_id);
CREATE INDEX idx_federation_peers_status ON federation_peers(status);
CREATE INDEX idx_federation_equipment_shares_tenant ON federation_equipment_shares(tenant_id);
CREATE INDEX idx_federation_equipment_shares_peer ON federation_equipment_shares(peer_id);
CREATE INDEX idx_federation_share_requests_tenant ON federation_share_requests(tenant_id);
CREATE INDEX idx_federation_share_requests_status ON federation_share_requests(status);
