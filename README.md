# Terraform Provider Dkron

Provider for managing https://dkron.io/ jobs.

# Usage examples

```terraform
terraform {
  required_providers {
    dkron = {
      version = "0.2"
      source  = "registry.terraform.io/bozerkins/dkron"
    }
  }
}

provider "dkron" {
  host = "http://localhost:8080"
}

resource "dkron_job" "job1" {
  name              = "job1"
  timezone          = "Europe/Riga"
  executor          = "shell"
  command           = "date"
  env               = "EDITOR=vi"
  cwd               = ""
  shell             = false
  allowed_exitcodes = "0, 199, 255"
  schedule          = "@every 10m"
  timeout           = "9s"
  mem_limit_kb      = "16384"
  project           = "myproject"
  disabled          = false
  retries           = 5
  concurrency       = "forbid"

  tags              = {
    "myproject" = "dumdum:2"
  }


  # out to fluent
  processors  {
    type = "fluent"
  }
}

resource "dkron_job" "job2" {
  name              = "job2"
  timezone          = "Europe/Riga"
  executor          = "shell"
  command           = "date"
  env               = "ENV1=envone"
  cwd               = ""
  shell             = false
  allowed_exitcodes = "0, 199, 255"
  schedule          = "@every 10m"
  timeout           = "9s"
  mem_limit_kb      = "16384"
  project           = "myproject"
  disabled          = false
  retries           = 5
  concurrency       = "forbid"

  tags              = {
    "myproject" = "dumdum"
  }

  processors  {
    type = "log"
  }
}
```

## Usability notes
Dkron API is poorly documented and doesn't always work as in documentation. Because of that some inconsistencies may arise

## Known issues
Processors order doesn't really work, so using more than one processor isn't advised. 
This happens because of how Go serializes JSON and how Dkron Job API works.

# Acknoledgments
This provider is based on https://github.com/peertransfer/terraform-provider-dkron provider

Special thank to https://github.com/andreygolev for improving the initial provider

---
# Terraform Provider Dkron

Run the following command to build the provider

```shell
go build -o terraform-provider-dkron
```

## Test sample configuration

First, build and install the provider.

```shell
make install
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```