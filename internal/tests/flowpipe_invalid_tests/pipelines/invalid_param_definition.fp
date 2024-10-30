pipeline "update_pubsub_topics" {
  param "application_credentials" {
    type        = "string"
    default     = "abc"
  }

  step "transform" "value" {
    value = param.application_credentials
  }
}