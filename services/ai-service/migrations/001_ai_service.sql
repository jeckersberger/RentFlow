-- Create AI Providers table
CREATE TABLE ai_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,          -- claude, gpt4o, gemini, mistral, ollama
    api_endpoint VARCHAR(500),
    model_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    priority INT DEFAULT 0,             -- fallback order
    config JSONB DEFAULT '{}',          -- provider-specific config
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

-- Create AI Requests table
CREATE TABLE ai_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    provider_id UUID NOT NULL REFERENCES ai_providers(id) ON DELETE CASCADE,
    request_type VARCHAR(50) NOT NULL,  -- price_optimization, demand_forecast, asset_creator, predictive_maintenance, general
    input_text TEXT NOT NULL,
    anonymized_input TEXT,
    output_text TEXT,
    model_used VARCHAR(100),
    tokens_used INT,
    latency_ms INT,
    status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT valid_status CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    CONSTRAINT valid_request_type CHECK (request_type IN ('price_optimization', 'demand_forecast', 'asset_creator', 'predictive_maintenance', 'general'))
);

-- Create AI Feedback table
CREATE TABLE ai_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    request_id UUID NOT NULL REFERENCES ai_requests(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    is_correct BOOLEAN,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create Few-Shot Examples table
CREATE TABLE few_shot_examples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    request_type VARCHAR(50) NOT NULL,  -- price_optimization, demand_forecast, asset_creator, predictive_maintenance, general
    input_example TEXT NOT NULL,
    output_example TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    usage_count INT DEFAULT 0,
    avg_rating FLOAT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT valid_request_type CHECK (request_type IN ('price_optimization', 'demand_forecast', 'asset_creator', 'predictive_maintenance', 'general'))
);

-- Create indices for performance
CREATE INDEX idx_ai_requests_tenant_id ON ai_requests(tenant_id);
CREATE INDEX idx_ai_requests_provider_id ON ai_requests(provider_id);
CREATE INDEX idx_ai_requests_created_at ON ai_requests(created_at DESC);
CREATE INDEX idx_ai_requests_status ON ai_requests(status);

CREATE INDEX idx_ai_feedback_tenant_id ON ai_feedback(tenant_id);
CREATE INDEX idx_ai_feedback_request_id ON ai_feedback(request_id);

CREATE INDEX idx_few_shot_examples_tenant_id ON few_shot_examples(tenant_id);
CREATE INDEX idx_few_shot_examples_request_type ON few_shot_examples(request_type, is_active);

CREATE INDEX idx_ai_providers_tenant_id ON ai_providers(tenant_id);
CREATE INDEX idx_ai_providers_active ON ai_providers(tenant_id, is_active);
