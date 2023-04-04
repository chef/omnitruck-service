# Omnitruck deployment 

Omnitruck wrapper services can be deployed manually and using Terraform script.


``` bash
Terraform :
`cd terraform`
`terraform plan`
```


``` bash
Manually :
1. Download omnitruck-service image from j-frog using below command
   `jfrog rt dl --apikey "$ARTIFACTORY_TOKEN"  https://artifactory-internal.ps.chef.co/artifactory/docker-chefes-local/omnitruck-services/ omnitruck-service/`
2. Login to aws 
   `aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 862552916454.dkr.ecr.us-east-1.amazonaws.com`
3. Tag your downloaded image to ECR
   ` docker tag <image-id> 862552916454.dkr.ecr.us-east-1.amazonaws.com/omitruck-services-acceptance/omnitruck-service`
4. Push your image to ECR
   `docker push 862552916454.dkr.ecr.us-east-1.amazonaws.com/omitruck-services-acceptance/omnitruck-service`


#Steps to generate Artifactory token:
1. export VAULT_ADDR="https://vault.ps.chef.co"
2. export ARTIFACTORY_TOKEN=$(vault kv get -field token <ARTIFACTORY-PATH>)
```