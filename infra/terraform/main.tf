provider "google" {
  project     = var.project_id
  region      = var.region
  credentials = file(var.credentials_file)
}

data "template_file" "startup" {
  template = file("${path.module}/cloud-init.sh")
  vars = {
    repo_url    = var.repo_url
    repo_branch = var.repo_branch
  }
}

resource "google_compute_instance" "dev_vm" {
  name         = lower("dev-env-${var.stack}-${var.instance_id}")
  machine_type = var.machine_type
  zone         = var.zone

  boot_disk {
    initialize_params {
      image = var.image
      size  = var.disk_size
    }
  }

  network_interface {
    network       = "default"
    access_config {}
  }

  metadata_startup_script = data.template_file.startup.rendered

  tags = ["dev-env"]

  labels = {
    stack      = var.stack
    owner      = lower(var.user_id)
    instanceid = var.instance_id
    created    = substr(timestamp(), 0, 10)
  }
}

output "instance_name" {
  value = google_compute_instance.dev_vm.name
}

output "instance_ip" {
  value = google_compute_instance.dev_vm.network_interface[0].access_config[0].nat_ip
}