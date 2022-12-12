locals {
  name        = "omnitruck-services"
  environment = "acceptance"
  full_name   = "${local.name}-${local.environment}"
  ecs_app_def = templatefile("./templates/ecs/app.json.tpl", {
    name           = local.name
    app_image      = var.app_image
    app_port       = var.app_port
    fargate_cpu    = var.fargate_cpu
    fargate_memory = var.fargate_memory
    aws_region     = var.aws_region
  })
}
