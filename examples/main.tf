terraform {
  required_providers {
    ansible = {
      version = "~> 1.0.0"
      source  = "rubencosta/ansible"
    }
  }
}

resource "ansible_playbook" "playbook" {
  playbook = "playbook.yaml"
  inventory_hosts = [{
    name = my_server.ipv4_address
  }]
}
