locals {
  services = [
    {
      domain    = "chefdownload-trial.chef.io"
      origin    = "trial-production.downloads.chef.co"
      fastly_id = "2TbsWof5nga4BDQrr2QUx2"
    },
    {
      domain    = "chefdownload-community.chef.io"
      origin    = "opensource-production.downloads.chef.co"
      fastly_id = "TI6lSOqWRDU7as2vUCT101"
    },
    {
      domain    = "chefdownload-commerical.chef.io"
      origin    = "commercial-production.downloads.chef.co"
      fastly_id = "P7mCOa11pnY0Qv0RhBScf6"
    }
  ]
}

resource "fastly_service_vcl" "omnitruck-service-cdn" {
  for_each = { for s in local.services : s.domain => s }

  name = each.value.domain

  domain {
    name = each.value.domain
  }

  backend {
    name              = "Origin"
    address           = each.value.origin
    ssl_cert_hostname = each.value.origin
    ssl_sni_hostname  = each.value.origin
    use_ssl           = true
    port              = 443
  }

  healthcheck {
    name              = "Status"
    method            = "GET"
    host              = each.value.origin
    path              = "/status"
    expected_response = 200

    check_interval = 60000
  }
}
