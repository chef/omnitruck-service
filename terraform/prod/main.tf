locals {
  name        = "omnitruck"
  environment = "production"
  full_name   = "${local.name}-${local.environment}"
  ecs_app_def = templatefile("./templates/ecs/app.json.tpl", {
    name                = local.full_name
    app_image           = "${data.aws_ecr_repository.omnitruck.repository_url}:${var.app_version_tag}@${data.aws_ecr_image.omnitruck.image_digest}"
    app_trial_port      = var.app_trial_port
    app_os_port         = var.app_os_port
    app_commercial_port = var.app_commercial_port
    fargate_cpu         = var.fargate_cpu
    fargate_memory      = var.fargate_memory
    aws_region          = var.aws_region
  })
}
