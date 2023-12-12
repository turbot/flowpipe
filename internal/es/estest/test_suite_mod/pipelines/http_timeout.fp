pipeline "pipeline_with_http_timeout" {

  param "timeout_string" {
    type    = string
    default = "15ms"
  }

  param "timeout_number" {
    type    = number
    default = 15
  }

  step "http" "http_with_timeout_string" {
    url     = "http://localhost:7104/delay"
    timeout = "10ms"
  }

  step "http" "http_with_timeout_number" {
    url     = "http://localhost:7104/delay"
    timeout = 10
  }

  step "http" "http_with_timeout_string_unresolved" {
    url     = "http://localhost:7104/delay"
    timeout = param.timeout_string
  }

  step "http" "http_with_timeout_number_unresolved" {
    url     = "http://localhost:7104/delay"
    timeout = param.timeout_number
  }
}