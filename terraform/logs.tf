# logs.tf

# Set up CloudWatch group and log stream and retain logs for 30 days
resource "aws_cloudwatch_log_group" "la_log_group" {
  name              = "/ecs/${local.name}"
  retention_in_days = 30

  tags = {
    Name = "${local.name}-log-group"
  }
}

resource "aws_cloudwatch_log_stream" "la_log_stream" {
  name           = "${local.name}-log-stream"
  log_group_name = aws_cloudwatch_log_group.la_log_group.name
}
