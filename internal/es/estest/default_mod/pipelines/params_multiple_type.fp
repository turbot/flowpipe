pipeline "tag_resources" {
  title       = "Tag Resources"
  description = "Applies one or more tags to the specified resources."

  tags = {
    type = "featured"
  }

  param "region" {
    type        = string
  }

  param "cred" {
    type        = string
    default     = "default"
  }

  param "resource_arns" {
    type        = list(string)
    description = "Specifies the list of ARNs of the resources that you want to apply tags to."
  }

  param "tags" {
    type        = map(string)
    description = "Specifies the tags that you want to apply to the specified resources."
  }

  step "transform" "echo" {
    value = merge(credential.jira[param.cred].env, { AWS_REGION = param.region })
  }

  output "region" {
    value = param.region
  }

  output "cred" {
    value = param.cred
  }

  output "resource_arns" {
    value = param.resource_arns
  }

  output "echo" {
    value = step.transform.echo
  }
}


