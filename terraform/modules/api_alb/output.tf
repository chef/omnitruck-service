output "alb_arn" {
    value = aws_alb.api.arn
}

output "alb_tg_arn" {
    value = aws_alb_target_group.api.arn
}

output "dns_name" {
    value = aws_alb.api.dns_name
}