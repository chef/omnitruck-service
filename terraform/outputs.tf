output "alb_trial_hostname" {
  value = aws_alb.trial.dns_name
}

output "alb_os_hostname" {
  value = aws_alb.opensource.dns_name
}

output "alb_comm_hostname" {
  value = aws_alb.commercial.dns_name
}
