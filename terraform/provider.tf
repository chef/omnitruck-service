
# provider.tf

# Specify the provider and access details
provider "vault" {
  address = "https://vault.ps.chef.co"
}

provider "aws" {
  profile = "chef-telemetry"
  region  = var.aws_region

  access_key = data.vault_aws_access_credentials.telemetry.access_key
  secret_key = data.vault_aws_access_credentials.telemetry.secret_key
  token      = data.vault_aws_access_credentials.telemetry.security_token
}

provider "aws" {
  alias  = "secure"
  region = var.aws_region

  access_key = data.vault_aws_access_credentials.secure.access_key
  secret_key = data.vault_aws_access_credentials.secure.secret_key
  token      = data.vault_aws_access_credentials.secure.security_token
}
