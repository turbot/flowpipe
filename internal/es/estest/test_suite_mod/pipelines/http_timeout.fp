pipeline "pipeline_with_http_timeout" {

  param "timeout_string" {
    type    = string
    default = "1ms"
  }

  param "timeout_number" {
    type    = number
    default = 100
  }
  
  step "http" "http_with_timeout_string" {
    url     = "https://steampipe.io"
    timeout = "1ms"
  }

  step "http" "http_with_timeout_number" {
    url     = "https://steampipe.io"
    timeout = 100
  }

  step "http" "http_with_timeout_string_unresolved" {
    url     = "https://steampipe.io"
    timeout = param.timeout_string
  }

  step "http" "http_with_timeout_number_unresolved" {
    url     = "https://steampipe.io"
    timeout = param.timeout_number
  }
}