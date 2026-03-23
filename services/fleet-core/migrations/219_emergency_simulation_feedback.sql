-- Add simulation and feedback fields to sos_incidents
ALTER TABLE sos_incidents ADD COLUMN IF NOT EXISTS simulation_mode BOOLEAN DEFAULT FALSE;
ALTER TABLE sos_incidents ADD COLUMN IF NOT EXISTS driver_feedback VARCHAR(50);
ALTER TABLE sos_incidents ADD COLUMN IF NOT EXISTS feedback_notes TEXT;
