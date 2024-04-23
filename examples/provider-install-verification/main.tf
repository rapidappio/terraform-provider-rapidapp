terraform {
  required_providers {
    rapidapp = {
      source = "registry.terraform.io/rapidappio/rapidapp"
    }
  }
}

provider "rapidapp" {}

data "rapidapp_postgres_databases" "example" {}

