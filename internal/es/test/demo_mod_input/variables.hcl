
variable "aws_region" {
  type        = string
  description = "AWS Region"
}

variable "aws_access_key_id" {
  type        = string
  description = "AWS Access Key ID"
}

variable "aws_secret_access_key" {
  type        = string
  description = "AWS Secret Access Key"
}

variable "slack_token" {
  type        = string
  description = "AWS Secret Access Key"
}

variable "smtp_username" {
  type        = string
  description = "SMTP email username"
}

variable "smtp_password" {
  type        = string
  description = "SMTP email Password"
}

variable "smtp_server" {
  type        = string
  description = "SMTP Host DNS name"
}

variable "smtp_port" {
  type        = string
  description = "SMTP Port"
}

variable "smtp_from" {
  type        = string
  description = "Sender address (From)"
}
variable "response_url" {
  type        = string
  description = "Publicly accessible endpoint for email callback"
}

// variable "openai_token" {
//   type        = string
//   description = "OpenAI API Token"
// }

locals {
  aws_creds_vars = {
    AWS_REGION              = var.aws_region
    AWS_ACCESS_KEY_ID       = var.aws_access_key_id
    AWS_SECRET_ACCESS_KEY   = var.aws_secret_access_key   
  }
}
