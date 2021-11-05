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

resource "dkron_job" "job1_dms" {
  name              = "job1_dms_1"
  timezone          = "Europe/Riga"
  owner             = "Gitlab"
  owner_email       = "gitlab@gitlabovich.com"
  executor          = "shell"
  command           = "date"
  cwd               = ""
  shell             = false
  allowed_exitcodes = "0, 199, 255"
  schedule          = "@every 10m"
  timeout           = "9s"
  mem_limit_kb      = "16384"
  project           = "dms"
  disabled          = false
  retries           = 5
  concurrency       = "forbid"

  tags              = {
    "dms" = "cron:1"
  }


  # out to fluent
  processors  {
    forward = "true"
    type = "fluent"
  }

  # output to stdin/stdou
  processors  {
    type = "log"
  }
}

resource "dkron_job" "job2_dms" {
  name              = "job1_dms_2"
  timezone          = "Europe/Riga"
  owner             = "Gitlab"
  owner_email       = "gitlab@gitlabovich.com"
  executor          = "shell"
  command           = "date"
  cwd               = ""
  shell             = false
  allowed_exitcodes = "0, 199, 255"
  schedule          = "@every 10m"
  timeout           = "9s"
  mem_limit_kb      = "16384"
  project           = "dms"
  disabled          = false
  retries           = 5
  concurrency       = "forbid"

  tags              = {
    "dms" = "cron:1"
  }

  processors  {
    type = "log"
  }
}