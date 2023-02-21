# alb.tf

module "trial_alb" {
  source = "../modules/api_alb"
  name = "omnitruck-accept-trial"
  environment = local.environment
  api_dns_name = "trial-acceptance.downloads.chef.co"
  api_port = 3001
  vpc_id = aws_vpc.main.id
  subnet_ids = aws_subnet.public.*.id 
  security_groups = [aws_security_group.lb.id]
  health_check_path = "/status"
}

module "opensource_alb" {
  source = "../modules/api_alb"
  name = "omnitruck-accept-os"
  environment = local.environment
  api_dns_name = "opensource-acceptance.downloads.chef.co"
  api_port = 3000
  vpc_id = aws_vpc.main.id
  subnet_ids = aws_subnet.public.*.id 
  security_groups = [aws_security_group.lb.id]
  health_check_path = "/status"
}

module "commercial_alb" {
  source = "../modules/api_alb"
  name = "omnitruck-accept-com"
  environment = local.environment
  api_dns_name = "commercial-acceptance.downloads.chef.co"
  api_port = 3002
  vpc_id = aws_vpc.main.id
  subnet_ids = aws_subnet.public.*.id 
  security_groups = [aws_security_group.lb.id]
  health_check_path = "/status"
}
