---
page_title: "Ansible Provider"
subcategory: ""
description: |-
  Terraform provider for Ansible.
---

# Ansible Provider

The Ansible provider is used to interact with Ansible.

Use the navigation to the left to read about the available resources.


## Example Usage

```terraform
terraform {
  required_providers {
    ansible = {
      version = "~> 1.0.0"
      source  = "rubencosta/ansible"
    }
  }
}
```
