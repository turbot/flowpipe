pipeline "update_pubsub_topics" {
  param "application_credentials" {
    type        = credential
    default     = "abc"
  }

  step "transform" "value" {
    value = param.application_credentials
  }
}