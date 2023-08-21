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
  }]
  inventory_groups = [{
    name = "group_parent"
    children = [
      "group_a",
    ]
    variables = yamlencode({
      group_var_a = "Group variable"
    })
  }]
  extra_vars = yamlencode({
    ansible_config_file = "${path.module}/ansible.cfg"
  })
}
