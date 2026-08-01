ALTER TABLE template_versions
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE TRIGGER set_template_versions_updated_at
BEFORE UPDATE ON template_versions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
