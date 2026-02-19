-- Create predictive_models table
CREATE TABLE IF NOT EXISTS predictive_models (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    model_type VARCHAR(100) NOT NULL, -- regression, classification, clustering
    description TEXT,
    algorithm VARCHAR(100) NOT NULL, -- random_forest, neural_network, svm, etc.
    input_features JSONB NOT NULL DEFAULT '[]', -- list of feature names as JSON array
    output_variable VARCHAR(255),
    training_data_start TIMESTAMP WITH TIME ZONE NOT NULL,
    training_data_end TIMESTAMP WITH TIME ZONE NOT NULL,
    accuracy DECIMAL(5,4) CHECK (accuracy >= 0 AND accuracy <= 1),
    precision DECIMAL(5,4) CHECK (precision >= 0 AND precision <= 1),
    recall DECIMAL(5,4) CHECK (recall >= 0 AND recall <= 1),
    f1_score DECIMAL(5,4) CHECK (f1_score >= 0 AND f1_score <= 1),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, training, error
    trained_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create prediction_results table
CREATE TABLE IF NOT EXISTS prediction_results (
    id VARCHAR(36) PRIMARY KEY,
    model_id VARCHAR(36) NOT NULL REFERENCES predictive_models(id) ON DELETE CASCADE,
    prediction_input JSONB NOT NULL, -- input features used for prediction
    predicted_value JSONB, -- predicted outcome (can be any JSON type)
    confidence_score DECIMAL(5,4) CHECK (confidence_score >= 0 AND confidence_score <= 1) NOT NULL,
    probability_scores JSONB, -- for classification models
    feature_importance JSONB, -- importance of each feature
    anomaly_detected BOOLEAN NOT NULL DEFAULT FALSE,
    anomaly_score DECIMAL(5,4) CHECK (anomaly_score >= 0 AND anomaly_score <= 1),
    explanation TEXT, -- explanation of prediction
    organization_id UUID NOT NULL,
    request_id VARCHAR(36) NOT NULL,
    predicted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create analytics_jobs table
CREATE TABLE IF NOT EXISTS analytics_jobs (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL, -- batch, real_time, streaming
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, running, completed, failed, cancelled
    job_type VARCHAR(100) NOT NULL, -- prediction, training, evaluation, inference
    parameters JSONB NOT NULL DEFAULT '{}', -- job-specific parameters
    input_source VARCHAR(255) NOT NULL, -- data source for the job
    output_destination VARCHAR(255) NOT NULL, -- destination for results
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    progress DECIMAL(5,2) CHECK (progress >= 0 AND progress <= 100) DEFAULT 0,
    error_details TEXT,
    organization_id UUID NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create fleet_analytics table
CREATE TABLE IF NOT EXISTS fleet_analytics (
    id VARCHAR(36) PRIMARY KEY,
    organization_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    total_vehicles INTEGER NOT NULL DEFAULT 0,
    active_vehicles INTEGER NOT NULL DEFAULT 0,
    total_distance DECIMAL(12,2) NOT NULL DEFAULT 0, -- in km
    average_speed DECIMAL(8,2) NOT NULL DEFAULT 0, -- in km/h
    total_fuel_consumption DECIMAL(12,2) NOT NULL DEFAULT 0, -- in liters
    average_fuel_efficiency DECIMAL(8,2) NOT NULL DEFAULT 0, -- km/liter
    total_trips INTEGER NOT NULL DEFAULT 0,
    completed_trips INTEGER NOT NULL DEFAULT 0,
    cancelled_trips INTEGER NOT NULL DEFAULT 0,
    average_trip_duration DECIMAL(8,2) NOT NULL DEFAULT 0, -- in hours
    idle_time_percentage DECIMAL(5,2) CHECK (idle_time_percentage >= 0 AND idle_time_percentage <= 100) NOT NULL DEFAULT 0,
    routing_efficiency DECIMAL(5,2) CHECK (routing_efficiency >= 0 AND routing_efficiency <= 100) NOT NULL DEFAULT 0, -- percentage improvement over baseline
    fuel_cost DECIMAL(12,2) NOT NULL DEFAULT 0, -- in currency
    maintenance_cost DECIMAL(12,2) NOT NULL DEFAULT 0, -- in currency
    revenue DECIMAL(15,2) NOT NULL DEFAULT 0, -- in currency
    profit_margin DECIMAL(5,2) NOT NULL DEFAULT 0, -- percentage
    safety_score DECIMAL(5,2) CHECK (safety_score >= 0 AND safety_score <= 100) NOT NULL DEFAULT 0, -- safety rating
    accidents_count INTEGER NOT NULL DEFAULT 0,
    incidents_count INTEGER NOT NULL DEFAULT 0,
    emergency_stops INTEGER NOT NULL DEFAULT 0,
    geofence_violations INTEGER NOT NULL DEFAULT 0,
    alerts_triggered INTEGER NOT NULL DEFAULT 0,
    data_points_collected BIGINT NOT NULL DEFAULT 0,
    processing_time DECIMAL(10,3) NOT NULL DEFAULT 0, -- in seconds
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create driver_behavior_analytics table
CREATE TABLE IF NOT EXISTS driver_behavior_analytics (
    id VARCHAR(36) PRIMARY KEY,
    driver_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    analysis_period_start DATE NOT NULL,
    analysis_period_end DATE NOT NULL,
    total_driving_hours DECIMAL(8,2) NOT NULL DEFAULT 0,
    aggressive_acceleration_count INTEGER NOT NULL DEFAULT 0,
    hard_braking_count INTEGER NOT NULL DEFAULT 0,
    sharp_turn_count INTEGER NOT NULL DEFAULT 0,
    over_speeding_count INTEGER NOT NULL DEFAULT 0,
    idling_time DECIMAL(8,2) NOT NULL DEFAULT 0, -- in hours
    average_speed DECIMAL(8,2) NOT NULL DEFAULT 0, -- in km/h
    max_speed DECIMAL(8,2) NOT NULL DEFAULT 0, -- in km/h
    fuel_efficiency DECIMAL(8,2) NOT NULL DEFAULT 0, -- km/liter
    safety_score DECIMAL(5,2) CHECK (safety_score >= 0 AND safety_score <= 100) NOT NULL DEFAULT 0, -- 0-100 scale
    efficiency_score DECIMAL(5,2) CHECK (efficiency_score >= 0 AND efficiency_score <= 100) NOT NULL DEFAULT 0, -- 0-100 scale
    comfort_score DECIMAL(5,2) CHECK (comfort_score >= 0 AND comfort_score <= 100) NOT NULL DEFAULT 0, -- 0-100 scale
    ranking INTEGER, -- rank among drivers
    improvement_suggestions JSONB NOT NULL DEFAULT '[]', -- suggestions for improvement
    behavior_trend VARCHAR(20) NOT NULL DEFAULT 'stable', -- improving, declining, stable
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create predictive_maintenance_analytics table
CREATE TABLE IF NOT EXISTS predictive_maintenance_analytics (
    id VARCHAR(36) PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    analysis_date DATE NOT NULL,
    next_service_due DATE NOT NULL,
    service_priority VARCHAR(20) NOT NULL, -- high, medium, low
    component_risks JSONB NOT NULL DEFAULT '{}', -- risk scores for each component
    failure_prediction JSONB, -- predicted failures
    maintenance_schedule JSONB NOT NULL DEFAULT '[]', -- recommended schedule
    parts_needed JSONB NOT NULL DEFAULT '[]', -- parts that may be needed
    cost_estimation DECIMAL(12,2) NOT NULL DEFAULT 0, -- estimated cost in currency
    uptime_prediction DECIMAL(5,2) CHECK (uptime_prediction >= 0 AND uptime_prediction <= 100) NOT NULL DEFAULT 0, -- predicted uptime percentage
    reliability_score DECIMAL(5,2) CHECK (reliability_score >= 0 AND reliability_score <= 100) NOT NULL DEFAULT 0, -- reliability rating
    recommendations JSONB NOT NULL DEFAULT '[]', -- maintenance recommendations
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_predictive_models_org_type_status ON predictive_models(organization_id, model_type, status);
CREATE INDEX IF NOT EXISTS idx_prediction_results_model_created ON prediction_results(model_id, created_at);
CREATE INDEX IF NOT EXISTS idx_prediction_results_org_created ON prediction_results(organization_id, created_at);
CREATE INDEX IF NOT EXISTS idx_prediction_results_anomaly ON prediction_results(anomaly_detected, created_at);
CREATE INDEX IF NOT EXISTS idx_analytics_jobs_org_status_type ON analytics_jobs(organization_id, status, job_type);
CREATE INDEX IF NOT EXISTS idx_fleet_analytics_org_dates ON fleet_analytics(organization_id, start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_driver_behavior_org_dates ON driver_behavior_analytics(organization_id, analysis_period_start, analysis_period_end);
CREATE INDEX IF NOT EXISTS idx_driver_behavior_driver ON driver_behavior_analytics(driver_id, analysis_period_start, analysis_period_end);
CREATE INDEX IF NOT EXISTS idx_pred_maint_org_vehicle ON predictive_maintenance_analytics(organization_id, vehicle_id);
CREATE INDEX IF NOT EXISTS idx_pred_maint_analysis_date ON predictive_maintenance_analytics(analysis_date);

-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update the updated_at timestamp
CREATE TRIGGER update_predictive_models_updated_at 
    BEFORE UPDATE ON predictive_models 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_prediction_results_updated_at 
    BEFORE UPDATE ON prediction_results 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_analytics_jobs_updated_at 
    BEFORE UPDATE ON analytics_jobs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_fleet_analytics_updated_at 
    BEFORE UPDATE ON fleet_analytics 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_driver_behavior_analytics_updated_at 
    BEFORE UPDATE ON driver_behavior_analytics 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_predictive_maintenance_analytics_updated_at 
    BEFORE UPDATE ON predictive_maintenance_analytics 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();