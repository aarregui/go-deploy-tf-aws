locals {
  identifier = "${var.environment}-${var.app_name}"

  api_host = "api.%{if var.environment != "prod"}${var.environment}.%{endif}${var.domain_name}"
}
