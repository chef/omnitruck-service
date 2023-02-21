 data "aws_route53_zone" "downloads" {
    name = "downloads.chef.co."
 }

 resource "aws_route53_record" "api" {
    zone_id = data.aws_route53_zone.downloads.zone_id
    name = var.api_dns_name
    type = "A"

    alias {
        name = aws_alb.api.dns_name
        zone_id = aws_alb.api.zone_id
        evaluate_target_health = true
    }
 }