package services

import (
	"fmt"
	"os"
	"os/exec"
)

// ExecuteTerraform runs terraform init, plan, and apply in the given per-instance directory.
func ExecuteTerraform(tfDir string) error {
	Info("Terraform directory: " + tfDir)

	// terraform init
	Info("Initializing Terraform...")
	cmdInit := exec.Command("terraform", "init")
	cmdInit.Dir = tfDir
	cmdInit.Stdout = os.Stdout
	cmdInit.Stderr = os.Stderr
	if err := cmdInit.Run(); err != nil {
		Error("terraform init failed")
		return fmt.Errorf("terraform init failed: %v", err)
	}
	Success("Terraform init completed")

	// terraform plan
	Info("Planning Terraform changes...")
	cmdPlan := exec.Command("terraform", "plan", "-var-file=terraform.tfvars")
	cmdPlan.Dir = tfDir
	cmdPlan.Stdout = os.Stdout
	cmdPlan.Stderr = os.Stderr
	if err := cmdPlan.Run(); err != nil {
		Error("terraform plan failed")
		return fmt.Errorf("terraform plan failed: %v", err)
	}
	Success("Terraform plan completed")

	// terraform apply
	Info("Applying Terraform changes...")
	cmdApply := exec.Command("terraform", "apply", "-auto-approve", "-var-file=terraform.tfvars")
	cmdApply.Dir = tfDir
	cmdApply.Stdout = os.Stdout
	cmdApply.Stderr = os.Stderr
	if err := cmdApply.Run(); err != nil {
		Error("terraform apply failed")
		return fmt.Errorf("terraform apply failed: %v", err)
	}
	Success("Terraform apply completed")

	return nil
}

// DestroyInfrastructure runs terraform destroy in the given per-instance directory.
func DestroyInfrastructure(tfDir string) error {
	Info("Destroying Terraform infrastructure...")
	cmd := exec.Command("terraform", "destroy", "-auto-approve", "-var-file=terraform.tfvars")
	cmd.Dir = tfDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("destroy failed: %v", err)
	}
	Success("Terraform destroy completed")

	return nil
}
