DROP TRIGGER IF EXISTS set_template_versions_updated_at ON template_versions;

ALTER TABLE template_versions
DROP COLUMN IF EXISTS updated_at;
