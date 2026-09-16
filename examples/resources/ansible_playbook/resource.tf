resource "ansible_playbook" "playbook" {
  replayable = true
  playbook   = "playbook.yaml"
  inventory_hosts = [{
    name   = my_server.ipv4_address
    groups = ["group_a"]
    variables = yamlencode({
      ansible_user = "admin"
      var_a        = "Host specific variable"
      var_b = [{
        nested_object_property = "Works"
      }]
    })
    # Never written to Terraform state. Requires Terraform >= 1.11.
    variables_wo = yamlencode({
      ansible_become_password = var.become_password
    })
  }]
  inventory_groups = [{
    name = "group_parent"
    children = [
      "group_a",
    ]
    variables = yamlencode({
      group_var_a = "Group variable"
    })
    variables_wo = yamlencode({
      vault_token = var.vault_token
    })
  }]
  extra_vars = yamlencode({
    ansible_config_file = "${path.module}/ansible.cfg"
  })
  extra_vars_wo = yamlencode({
    api_token = var.api_token
  })
}
