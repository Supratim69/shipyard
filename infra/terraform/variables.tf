variable "project_id" {}
variable "region" {
  default = "asia-south1"
}
variable "zone" {
  default = "asia-south1-c"
}
variable "credentials_file" {
  description = "Path to the service account JSON key"
}
variable "machine_type" {
  default = "e2-medium"
}
variable "disk_size" {
  default = 20
}
variable "image" {
  default = "debian-cloud/debian-11"
}
variable "stack" {
  description = "Selected tech stack (e.g., python, nodejs)"
}
variable "user_id" {
  description = "GitHub username or user-specific tag"
}
variable "instance_id" {
  type = string
}
variable "repo_url" {
  type = string
  description = "GitHub repository to clone"
}
variable "repo_branch" {
  type = string
  default = "main"
  description = "Branch to checkout"
}