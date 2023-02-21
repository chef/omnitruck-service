output "alb_trial_hostname" {
  value = module.trial_alb.dns_name
}

output "alb_os_hostname" {
  value = module.opensource_alb.dns_name
}

output "alb_comm_hostname" {
  value = module.commercial_alb.dns_name
}
