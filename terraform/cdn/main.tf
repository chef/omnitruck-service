terraform {
  required_providers {
    fastly = {
      source  = "fastly/fastly"
      version = ">= 4.2.0"
    }
  }

  # Requires AWS_ACCESS_KEY_ID environment variable
  backend "s3" {
    bucket  = "chef-telemetry-terraform-state"
    key     = "omnitruck-service/cdn.tfstate"
    region  = "us-west-2"
    # use aws' built-in at-rest encryption
    encrypt = true
    # TODO: consider kms_key_id field or AWS_SSE_CUSTOMER_KEY env var to
    # more deliberately manage state file encryption.
  }
}

# Requires FASTLY_API_KEY environment variable
provider "fastly" {
}
