# RentFlow Monitoring Stack

This directory contains all monitoring configurations for the RentFlow platform using Prometheus, Grafana, Loki, and Alertmanager.

## Directory Structure

```
monitoring/
├── prometheus/
│   └── alerts.yml              # Prometheus alerting rules
├── alertmanager/
│   └── alertmanager.yml        # Alertmanager configuration
├── grafana/
│   ├── dashboards/             # JSON dashboard templates
│   │   ├── service-latency.json
│   │   ├── error-rate.json
│   │   └── event-throughput.json
│   └── provisioning/
│       ├── datasources/        # Datasource definitions
│       │   └── datasources.yml
│       └── dashboards/         # Dashboard provisioning config
│           └── dashboards.yml
├── loki/
│   └── loki-config.yml         # Loki log aggregation config
└── promtail/
    └── promtail-config.yml     # Promtail agent config for log collection
```

## Components

### 1. Prometheus Alerting Rules (`prometheus/alerts.yml`)
Defines alert conditions for the monitoring stack:
- **ServiceDown**: Detects when any RentFlow service is unavailable (critical)
- **HighErrorRate**: Triggers when 5xx error rate exceeds 5% (warning)
- **HighLatency**: Triggers when p95 latency exceeds 2 seconds (warning)
- **PostgresConnectionPoolExhaustion**: Database connection pool > 80% (warning)
- **RedisHighMemoryUsage**: Redis memory > 80% of max (warning)
- **KurrentDBLowDiskSpace**: Event store has < 10% free space (critical)

**Integration**: Reference updated in `/infra/prometheus/prometheus.yml`

### 2. Alertmanager (`alertmanager/alertmanager.yml`)
Routes and manages alert notifications:
- Groups alerts by alertname, cluster, and service
- Separates critical alerts (10s wait) from normal alerts (30s wait)
- Sends webhooks to alert handler service (localhost:5001)
- Implements inhibition rules to suppress lower-severity duplicates

### 3. Grafana Dashboards

#### Service Latency Dashboard (`grafana/dashboards/service-latency.json`)
Visualizes HTTP request performance metrics:
- P50, P95, P99 latency percentiles by service
- Request rate per service
- 30s refresh rate, 6-hour time window

#### Error Rate Dashboard (`grafana/dashboards/error-rate.json`)
Monitors application error patterns:
- 5xx error rate percentage by service (color-coded thresholds)
- 5xx error count per service
- Error distribution pie chart
- Overall 5xx error rate stat card

#### Event Throughput Dashboard (`grafana/dashboards/event-throughput.json`)
Tracks KurrentDB (EventStore) performance:
- Events written and read rates (events/sec)
- Event count per 5m interval (stacked bar chart)
- Current write and read rate stat cards
- 30s refresh rate

### 4. Loki Log Aggregation (`loki/loki-config.yml`)
Central log storage and querying:
- Uses BoltDB for indexing with filesystem storage
- 24-hour retention and rotation
- Supports memcached results caching
- Query alignment with 15s steps

### 5. Promtail Log Collection (`promtail/promtail-config.yml`)
Collects and ships Docker container logs:
- **Docker SD Config**: Discovers containers via Docker socket
- **Container Labels**: Extracts service, container name, and network info
- **Pipeline Stages**:
  - CRI log parsing for container runtime logs
  - Multiline pattern matching for log concatenation
  - Regex extraction of timestamp, level, and message
  - Timestamp parsing and UTC normalization

### 6. Grafana Provisioning

#### Datasources (`grafana/provisioning/datasources/datasources.yml`)
Pre-configured data sources:
- **Prometheus**: Primary metrics (default)
- **Loki**: Log aggregation
- **Postgres**: Direct database queries (optional)

#### Dashboards (`grafana/provisioning/dashboards/dashboards.yml`)
Auto-provisioning configuration:
- Loads dashboards from `/etc/grafana/provisioning/dashboards`
- Allows UI updates and deletes via Grafana UI
- Reloads every 10 seconds

## Running the Monitoring Stack

### Option 1: Combined Docker Compose
```bash
# Run main services + monitoring
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d
```

### Option 2: Separate Monitoring Stack
```bash
# Run monitoring services independently
docker-compose -f docker-compose.monitoring.yml up -d
```

## Accessing Services

| Service | URL | Default Credentials |
|---------|-----|-------------------|
| Grafana | `http://localhost:3001` | admin / admin |
| Prometheus | `http://localhost:9090` | - |
| Alertmanager | `http://localhost:9093` | - |
| Loki | `http://localhost:3100` (API) | - |

## Alerts Flow

1. **Metrics Collection**: Prometheus scrapes metrics from all services every 15 seconds
2. **Rule Evaluation**: Alert rules evaluated every 30 seconds
3. **Alert Routing**: Triggered alerts sent to Alertmanager
4. **Alert Grouping**: Alertmanager groups similar alerts
5. **Notification**: Routes to webhook receiver (`localhost:5001/alerts`)

## Metric Requirements

Services must expose Prometheus metrics at `/metrics` endpoint:
- **http_requests_total**: Total HTTP requests (with status code label)
- **http_request_duration_seconds**: HTTP request latency (histogram)
- **kurrentdb_events_written_total**: Events written to KurrentDB
- **kurrentdb_events_read_total**: Events read from KurrentDB
- **pg_stat_activity_count**: Active Postgres connections
- **pg_settings_max_connections**: Max Postgres connections
- **redis_memory_used_bytes**: Redis memory usage
- **redis_memory_max_bytes**: Redis max memory
- **node_filesystem_avail_bytes**: Available disk space
- **node_filesystem_size_bytes**: Total disk size

## Customization

### Add New Alert Rules
Edit `monitoring/prometheus/alerts.yml` and add rule under the appropriate group.

### Modify Dashboard
Edit JSON files in `monitoring/grafana/dashboards/` or create new dashboards via Grafana UI (exported as JSON).

### Change Alert Thresholds
Update expressions in `monitoring/prometheus/alerts.yml` (latency, error rate, etc.)

### Update Alertmanager Routes
Edit `monitoring/alertmanager/alertmanager.yml` to change notification endpoints or routing logic.

## Storage Volumes

- **grafana_data**: Grafana configurations and provisioned dashboards
- **loki_data**: Loki index and chunk storage
- **alertmanager_data**: Alertmanager state and notification history
- **prometheus_data**: Time-series database (30 days retention by default)

## Environment Variables

Required in `.env` file:
- `GRAFANA_PASSWORD`: Admin password for Grafana (default: admin)
- `PROMETHEUS_RETENTION`: Data retention period (default: 30d)

## Architecture Notes

- **Prometheus**: Scrapes metrics directly from service `/metrics` endpoints
- **Loki**: Aggregates logs from Docker containers via Promtail
- **Grafana**: Unified dashboard platform querying both Prometheus and Loki
- **Alertmanager**: Handles alert routing and deduplication
- **Separate Network**: Monitoring services use the `rentflow` network for service discovery

This monitoring setup provides comprehensive observability for the RentFlow platform covering metrics, logs, and alerting.
