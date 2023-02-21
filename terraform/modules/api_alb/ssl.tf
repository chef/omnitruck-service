resource "aws_acm_certificate" "api_ssl" {
  domain_name = aws_route53_record.api.fqdn
  validation_method = "DNS"

  tags = {
    Environment = var.environment
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "api_ssl_validation" {
  zone_id = data.aws_route53_zone.downloads.zone_id
  allow_overwrite = true 
  ttl = "60"
  for_each = {
    for dvo in aws_acm_certificate.api_ssl.domain_validation_options : dvo.domain_name => {
        name = dvo.resource_record_name 
        type = dvo.resource_record_type 
        record = dvo.resource_record_value
    }
  }
  name = each.value.name 
  type = each.value.type 
  records = [ each.value.record ]
}