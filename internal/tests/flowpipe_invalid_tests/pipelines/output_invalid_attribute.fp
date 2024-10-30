pipeline "pipeline_output_with_invalid_attribute" {

  description = "Pipelines with a output block contains an invalid attrinute - sensitive"

  param "greetings" {
    type    = string
    default = "Hello world!"
  }

  output "greet_world" {
    description = "A greetings message."
    value       = param.greetings
    sensitive   = true
  }
}
