terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
    archive = {
      source = "hashicorp/archive"
    }
  }

  required_version = ">= 1.12.0"
}

provider "aws" {
  region = "eu-north-1"

  default_tags {
    tags = {
      app = "email-sub"
    }
  }
}
