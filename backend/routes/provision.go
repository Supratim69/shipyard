package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"oneclickdevenv/backend/db"
	"oneclickdevenv/backend/models"
	"oneclickdevenv/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProvisionRequest struct {
	UserID     string `json:"user_id"`
	Stack      string `json:"stack"`
	Machine    string `json:"machine_type"`
	Zone       string `json:"zone"`
	Image      string `json:"image"`
	DiskSize   int    `json:"disk_size"`
	RepoURL    string `json:"repo_url"`
	RepoBranch string `json:"repo_branch"`
}

type ProvisionResponse struct {
	Message      string     `json:"message"`
	InstanceID   string     `json:"instance_id"`
	InstanceName string     `json:"instance_name"`
	InstanceIP   string     `json:"instance_ip"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func ProvisionVM(c *gin.Context) {
	services.Info("Received provisioning request")

	var req ProvisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		services.Error("Invalid JSON body")
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	services.Info("Parsed request body")

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		services.Error("Invalid user ID format")
		c.JSON(400, gin.H{"error": "Invalid user ID format"})
		return
	}

	instanceUUID := uuid.New()
	tfDir := filepath.Join("..", "infra", "terraform", userUUID.String()+"_"+instanceUUID.String())
	if err := os.MkdirAll(tfDir, os.ModePerm); err != nil {
		services.Error(fmt.Sprintf("Failed to create tfDir: %v", err))
		c.JSON(500, gin.H{"error": "Failed to create instance directory"})
		return
	}

	for _, fname := range []string{"main.tf", "variables.tf", "outputs.tf", "cloud-init.sh", "terraform-sa-key.json"} {
		src := filepath.Join("..", "infra", "terraform", fname)
		dst := filepath.Join(tfDir, fname)
		if err := copyFile(src, dst); err != nil {
			services.Error(fmt.Sprintf("Failed to copy %s: %v", fname, err))
			c.JSON(500, gin.H{"error": "Failed to copy Terraform files"})
			return
		}
	}

	tfVarsTemplate := `project_id = "one-click-dev-env"
region = "asia-south1"
zone = "{{.Zone}}"
machine_type = "{{.Machine}}"
image = "{{.Image}}"
disk_size = {{.DiskSize}}
user_id = "{{.UserID}}"
instance_id = "{{.InstanceID}}"
stack = "{{.Stack}}"
repo_url = "{{.RepoURL}}"
repo_branch = "{{.RepoBranch}}"
credentials_file = "terraform-sa-key.json"`

	tmpl, err := template.New("tfvars").Parse(tfVarsTemplate)
	if err != nil {
		services.Error("Template parse error")
		c.JSON(500, gin.H{"error": "Template parse error"})
		return
	}
	services.Info("Parsed tfvars template")

	tfVarsPath := filepath.Join(tfDir, "terraform.tfvars")
	tfVarsFile, err := os.Create(tfVarsPath)
	if err != nil {
		services.Error(fmt.Sprintf("Failed to create terraform.tfvars: %v", err))
		c.JSON(500, gin.H{"error": "File creation error"})
		return
	}
	defer tfVarsFile.Close()

	data := struct {
		UserID     string
		Stack      string
		Machine    string
		Zone       string
		Image      string
		DiskSize   int
		InstanceID string
		RepoURL    string
		RepoBranch string
	}{
		UserID:     req.UserID,
		Stack:      req.Stack,
		Machine:    req.Machine,
		Zone:       req.Zone,
		Image:      req.Image,
		DiskSize:   req.DiskSize,
		InstanceID: instanceUUID.String(),
		RepoURL:    req.RepoURL,
		RepoBranch: req.RepoBranch,
	}

	if err := tmpl.Execute(tfVarsFile, data); err != nil {
		services.Error("Failed to execute tfvars template")
		c.JSON(500, gin.H{"error": "Template execution error"})
		return
	}
	services.Success("terraform.tfvars file created")

	// Run terraform init/apply in the per-instance directory
	if err := services.ExecuteTerraform(tfDir); err != nil {
		services.Error("Terraform provisioning failed")
		c.JSON(500, gin.H{"error": "Terraform error: " + err.Error()})
		return
	}
	services.Success("Terraform provisioning completed ✅")

	// Fetch terraform outputs from the per-instance directory
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = tfDir
	outputJSON, err := cmd.Output()
	if err != nil {
		services.Error(fmt.Sprintf("Failed to get terraform output: %v", err))
		c.JSON(500, gin.H{"error": "Failed to fetch output"})
		return
	}

	var tfOutput map[string]struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(outputJSON, &tfOutput); err != nil {
		services.Error(fmt.Sprintf("Failed to parse terraform output: %v", err))
		c.JSON(500, gin.H{"error": "Output parsing failed"})
		return
	}

	name, nameOk := tfOutput["instance_name"]
	ip, ipOk := tfOutput["instance_ip"]
	if !nameOk || !ipOk {
		services.Error("Terraform output missing expected keys")
		c.JSON(500, gin.H{"error": "Output missing instance_name or instance_ip"})
		return
	}

	// Set default TTL (e.g., 4 hours)
	ttlHours := 4
	expiresAt := time.Now().Add(time.Duration(ttlHours) * time.Hour)

	instance := models.Instance{
		ID:          instanceUUID,
		UserID:      userUUID,
		Stack:       req.Stack,
		Zone:        req.Zone,
		MachineType: req.Machine,
		Image:       req.Image,
		DiskSize:    req.DiskSize,
		Name:        name.Value,
		IP:          ip.Value,
		ExpiresAt:   &expiresAt,
		Status:      "active",
	}

	if err := db.DB.Create(&instance).Error; err != nil {
		services.Error("Failed to save instance to DB: " + err.Error())
		c.JSON(500, gin.H{"error": "Failed to save instance"})
		return
	}
	services.Success("Instance saved to DB")

	resp := ProvisionResponse{
		Message:      "VM provisioned successfully",
		InstanceID:   instanceUUID.String(),
		InstanceName: name.Value,
		InstanceIP:   ip.Value,
		ExpiresAt:    &expiresAt,
	}

	c.JSON(200, resp)
}
