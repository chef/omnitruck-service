[
  {
    "name": "${name}-app",
    "image": "${app_image}",
    "cpu": ${fargate_cpu},
    "memory": ${fargate_memory},
    "networkMode": "awsvpc",
    "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/${name}-app",
          "awslogs-region": "${aws_region}",
          "awslogs-stream-prefix": "ecs",
          "awslogs-create-group": "true"
        }
    },
    "portMappings": [
      {
        "containerPort": ${app_trial_port},
        "hostPort": ${app_trial_port}
      },
      {
        "containerPort": ${app_os_port},
        "hostPort": ${app_os_port}
      },
      {
        "containerPort": ${app_commercial_port},
        "hostPort": ${app_commercial_port}
      }
    ]
  }
]
