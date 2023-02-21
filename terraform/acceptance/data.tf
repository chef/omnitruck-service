data "vault_aws_access_credentials" "telemetry" {
  type    = "sts"
  backend = "account/dynamic/aws/chef-telemetry"
  role    = "default"
}

data "vault_aws_access_credentials" "secure" {
  type    = "sts"
  backend = "account/dynamic/aws/chef-secure"
  role    = "default"
}

# Fetch AZs in the current region
data "aws_availability_zones" "available" {
}

# data "aws_route53_zone" "secure" {
#   provider     = aws.secure
#   name         = "chef.co."
#   private_zone = true
# }

data "aws_ecr_repository" "omnitruck" {
  name = "omitruck-services-acceptance/omnitruck-service"
}

data "aws_ecr_image" "omnitruck" {
  repository_name = "omitruck-services-acceptance/omnitruck-service"
  image_tag = "${var.app_version_tag}"
}