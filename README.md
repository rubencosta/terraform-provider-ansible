# Terraform Provider for Ansible

This started as a fork of the [official Ansible Terraform Provider](https://github.com/ansible/terraform-provider-ansible) that ended up in a almost full rewrite for my use case:

- Run an Ansible Playbook for many hosts that were created in Terraform and run it from Terraform

The Terraform Provider for Ansible provides a straightforward way to run an Ansible Playbook while having the Ansible Inventory provided by Terraform.

This provider can be [found in the Terraform Registry here](https://registry.terraform.io/providers/rubencosta/ansible/latest).

## Requirements

- install Go: [official installation guide](https://go.dev/doc/install)
- install Terraform: [official installation guide](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli)
- install Ansible: [official installation guide](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)

## Installation for Local Development

Run `make`. This will build a `terraform-provider-ansible` binary in the top level of the project. To get Terraform to use this binary, configure the [development overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) for the provider installation. The easiest way to do this will be to create a config file with the following contents:

```
provider_installation {
  dev_overrides {
    "rubencosta/ansible" = "/path/to/project/root"
  }

  direct {}
}
```

The `/path/to/project/root` should point to the location where you have cloned this repo, where the `terraform-provider-ansible` binary will be built. You can then set the `TF_CLI_CONFIG_FILE` environment variable to point to this config file, and Terraform will use the provider binary you just built.

### Testing

Lint:

```shell
curl -L https://github.com/golangci/golangci-lint/releases/download/v1.50.1/golangci-lint-1.50.1-linux-amd64.tar.gz \
    | tar --wildcards -xzf - --strip-components 1 "**/golangci-lint"
curl -L https://github.com/nektos/act/releases/download/v0.2.34/act_Linux_x86_64.tar.gz \
    | tar -xzf - act

# linters
./golangci-lint run -v

# tests
make test

# GH actions locally
./act
```

### Examples

The [examples](./examples/) subdirectory contains a usage example for this provider.

## Releasing

To release a new version of the provider:

1. Update the version number in https://github.com/ansible/terraform-provider-ansible/blob/main/examples/provider/provider.tf
2. Run `go generate` to regenerate docs
3. Commit changes
4. Push a new tag (this should trigger an automated release process to the Terraform Registry)
5. Verify the new version is published at https://registry.terraform.io/providers/rubencosta/ansible/latest

## Licensing

GNU General Public License v3.0. See [LICENSE](/LICENSE) for full text.
