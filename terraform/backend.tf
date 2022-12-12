terraform {
  required_version = "~> 1.0"

  backend "http" {
    address        = "https://expeditor.chef.io/api/terraform/repo/omnitruck-services/resource/omnitruck-services-acceptance-api"
    lock_address   = "https://expeditor.chef.io/api/terraform/repo/omnitruck-services/resource/omnitruck-services-acceptance-api/lock"
    unlock_address = "https://expeditor.chef.io/api/terraform/repo/omnitruck-services/resource/omnitruck-services-acceptance-api/lock"

    lock_method   = "POST"
    unlock_method = "DELETE"
  }
}