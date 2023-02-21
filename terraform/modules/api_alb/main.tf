resource "aws_alb" "api" {
  name            = "${var.name}-api-alb"
  subnets         = var.subnet_ids
  security_groups = var.security_groups
}

resource "aws_alb_target_group" "api" {
  name        = "${var.name}-api-tg"
  port        = var.api_port
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    healthy_threshold   = "3"
    interval            = "30"
    protocol            = "HTTP"
    matcher             = "200"
    timeout             = "3"
    path                = var.health_check_path
    unhealthy_threshold = "2"
  }
}

# Redirect all traffic from the ALB to the target group
resource "aws_alb_listener" "api" {
  load_balancer_arn = aws_alb.api.id
  port              = 80
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_alb_target_group.api.id
    type             = "forward"
  }
}

resource "aws_alb_listener" "api_https" {
  load_balancer_arn = aws_alb.api.id 
  port = 443
  protocol = "HTTPS"
  certificate_arn = aws_acm_certificate.api_ssl.arn

  default_action {
    target_group_arn = aws_alb_target_group.api.id 
    type = "forward"
  }
}