CREATE TABLE workflow_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    trigger_type VARCHAR(30) NOT NULL,
    trigger_config JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    is_template BOOLEAN DEFAULT false,
    template_category VARCHAR(50),
    version INT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_workflow_definitions_tenant_id ON workflow_definitions(tenant_id);
CREATE INDEX idx_workflow_definitions_is_template ON workflow_definitions(is_template);

CREATE TABLE workflow_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    definition_id UUID REFERENCES workflow_definitions(id) ON DELETE CASCADE,
    status VARCHAR(30) DEFAULT 'running',
    trigger_data JSONB DEFAULT '{}',
    context_data JSONB DEFAULT '{}',
    current_step_index INT DEFAULT 0,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error_message TEXT
);

CREATE INDEX idx_workflow_instances_tenant_id ON workflow_instances(tenant_id);
CREATE INDEX idx_workflow_instances_definition_id ON workflow_instances(definition_id);
CREATE INDEX idx_workflow_instances_status ON workflow_instances(status);

CREATE TABLE workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    definition_id UUID REFERENCES workflow_definitions(id) ON DELETE CASCADE,
    step_index INT NOT NULL,
    name VARCHAR(200) NOT NULL,
    action_type VARCHAR(30) NOT NULL,
    action_config JSONB DEFAULT '{}',
    on_success_step INT,
    on_failure_step INT,
    timeout_seconds INT DEFAULT 300,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(definition_id, step_index)
);

CREATE INDEX idx_workflow_steps_definition_id ON workflow_steps(definition_id);
