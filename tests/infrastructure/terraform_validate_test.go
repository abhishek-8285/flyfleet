package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestTerraformFormat validates that all Terraform files are properly formatted
func TestTerraformFormat(t *testing.T) {
	terraformDir := "../../infrastructure/terraform"

	// Check terraform fmt
	cmd := exec.Command("terraform", "fmt", "-check", "-recursive", terraformDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Terraform files are not formatted correctly. Run 'terraform fmt -recursive':\n%s", output)
	}
}

// TestTerraformInit ensures backend is configured and init succeeds
func TestTerraformInit(t *testing.T) {
	terraformDir := "../../infrastructure/terraform"

	// Skip if testing in environment without AWS credentials
	if os.Getenv("SKIP_TERRAFORM_TESTS") == "true" {
		t.Skip("Skipping Terraform init test (SKIP_TERRAFORM_TESTS set)")
	}

	// Check if backend config exists
	backendFile := filepath.Join(terraformDir, "backend.tf")
	if _, err := os.Stat(backendFile); os.IsNotExist(err) {
		t.Fatal("backend.tf is missing - required for remote state")
	}

	// Note: Actual terraform init requires backend credentials
	// This test just checks configuration structure
	t.Log("Terraform backend configuration exists")
}

// TestTerraformValidate checks HCL syntax
func TestTerraformValidate(t *testing.T) {
	terraformDir := "../../infrastructure/terraform"

	// Run terraform validate
	cmd := exec.Command("terraform", "validate", "-json")
	cmd.Dir = terraformDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Terraform validation failed:\n%s", output)
	}
	t.Log("Terraform configuration is valid")
}

// TestRequiredFiles checks that all mandatory files exist
func TestRequiredFiles(t *testing.T) {
	requiredFiles := []string{
		"../../infrastructure/terraform/variables.tf",
		"../../infrastructure/terraform/main-new.tf",
		"../../infrastructure/terraform/backend.tf",
		"../../infrastructure/terraform/modules/networking/main.tf",
		"../../infrastructure/terraform/modules/kubernetes/main.tf",
		"../../infrastructure/terraform/modules/database/main.tf",
		"../../infrastructure/terraform/modules/iam/main.tf",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Required file missing: %s", file)
		}
	}
}

// TestModuleInputsOutputs validates that modules have proper input/output definitions
func TestModuleInputsOutputs(t *testing.T) {
	modules := []string{
		"networking",
		"kubernetes",
		"database",
		"iam",
		"kms",
		"ecr",
	}

	for _, module := range modules {
		moduleDir := filepath.Join("../../infrastructure/terraform/modules", module)

		// Check variables.tf exists
		varsFile := filepath.Join(moduleDir, "variables.tf")
		if _, err := os.Stat(varsFile); os.IsNotExist(err) {
			t.Errorf("Module %s missing variables.tf", module)
		}

		// Check outputs.tf exists
		outputsFile := filepath.Join(moduleDir, "outputs.tf")
		if _, err := os.Stat(outputsFile); os.IsNotExist(err) {
			t.Errorf("Module %s missing outputs.tf", module)
		}
	}
}
