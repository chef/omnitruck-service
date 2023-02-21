# ecs.tf

resource "aws_ecs_cluster" "main" {
  name = "${local.full_name}-cluster"
}

resource "aws_ecs_task_definition" "app" {
  family                   = "${local.full_name}-task"
  execution_role_arn       = "arn:aws:iam::862552916454:role/ecsTaskExecutionRole"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.fargate_cpu
  memory                   = var.fargate_memory
  container_definitions    = local.ecs_app_def
}

resource "aws_ecs_service" "main" {
  name            = local.full_name
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = var.app_count
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.ecs_tasks.id]
    subnets          = aws_subnet.private.*.id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = module.trial_alb.alb_tg_arn
    container_name   = "${local.full_name}-app"
    container_port   = var.app_trial_port
  }

  load_balancer {
    target_group_arn = module.opensource_alb.alb_tg_arn
    container_name   = "${local.full_name}-app"
    container_port   = var.app_os_port
  }

  load_balancer {
    target_group_arn = module.commercial_alb.alb_tg_arn
    container_name   = "${local.full_name}-app"
    container_port   = var.app_commercial_port
  }

  lifecycle {
    ignore_changes = [desired_count]
  }
}

# depends_on = [aws_alb_listener.front_end]
