pipeline "invalid_http_timeout" {
  step "http" "http_test" {
    url     = "https://somerandomsite.com"
    timeout = [100]
  }
}
