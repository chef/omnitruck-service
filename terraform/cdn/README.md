# cdn

This directory defines the terraform configuration for the omnitruck service's Fastly CDN frontend. 

See Local Setup for how this can be deployed from a local system. See also `.expeditor/terraform-cdn.pipeline.yml` which is a test to see how this terraform configuration could be deployed via buildkite pipelines.

## Local Setup

Install the required terraform version identified in `main.tf`. Consider using [tfenv](https://github.com/tfutils/tfenv) for this purpose, which can manage terraform versions for multiple environments at once (similar to rbenv/pyenv/virtualenv etc) by running `tfenv install` or `tfenv install min-requested`.

Log in to AWS:

```bash
# Log in and select the chef-telemetry AWS account, configuring it as the AWS profile named chef-telemetry:
saml2aws login --profile chef-telemetry --force
export AWS_DEFAULT_PROFILE=chef-telemetry
export AWS_ACCESS_KEY_ID="$(aws --profile chef-telemetry configure get aws_access_key_id)"
```

Get a fastly api and put it into an env var. This can be acquired from account settings in the fastly webui, or by getting the existing api key in vault at `secret/fastly/data/omnitruck-service` (note: the api key in vault is explicitly limited to managing the three fastly service ids identified below):

```bash
export FASTLY_API_KEY="<your-api-key>"
# -- OR --
export FASTLY_API_KEY="$(vault read --field=data --format=json 'secret/fastly/data/omnitruck-service' | jq -r .token)"
```

## Usage

```bash
terraform init
terraform plan # just view expected changes

# If deploying for the first time, import the specific fastly service ids that the token is allowed to manage:
terraform import 'fastly_service_vcl.omnitruck-service-cdn["chefdownload-trial.chef.io"]' '2TbsWof5nga4BDQrr2QUx2'
terraform import 'fastly_service_vcl.omnitruck-service-cdn["chefdownload-community.chef.io"]' 'TI6lSOqWRDU7as2vUCT101'
terraform import 'fastly_service_vcl.omnitruck-service-cdn["chefdownload-commerical.chef.io"]' 'P7mCOa11pnY0Qv0RhBScf6'

# Deploy by saving the plan in a file then applying it
terraform plan -out=default.tfplan
terraform apply default.tfplan
```
