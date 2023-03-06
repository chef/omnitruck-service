# alb.tf

module "trial_alb" {
  source = "../modules/api_alb"
  name = "omnitruck-prod-trial"
  environment = local.environment
  api_dns_name = "trial-production.downloads.chef.co"
  api_port = 3001
  vpc_id = aws_vpc.main.id
  subnet_ids = aws_subnet.public.*.id 
  security_groups = [aws_security_group.lb.id]
  health_check_path = "/status"
}

module "opensource_alb" {
  source = "../modules/api_alb"
  name = "omnitruck-prod-os"
  environment = local.environment
  api_dns_name = "opensource-production.downloads.chef.co"
  api_port = 3000
  vpc_id = aws_vpc.main.id
  subnet_ids = aws_subnet.public.*.id 
  security_groups = [aws_security_group.lb.id]
  health_check_path = "/status"
}

module "commercial_alb" {
  source = "../modules/api_alb"
  name = "omnitruck-prod-com"
  environment = local.environment
  api_dns_name = "commercial-production.downloads.chef.co"
  api_port = 3002
  vpc_id = aws_vpc.main.id
  subnet_ids = aws_subnet.public.*.id 
  security_groups = [aws_security_group.lb.id]
  health_check_path = "/status"
}
