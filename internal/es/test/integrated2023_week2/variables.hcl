variable "token" {
  type        = string
  description = "The GitHub personal access token to authenticate to the GitHub APIs, e.g., `github_pat_a1b2c3d4e5f6g7h8i9j10k11l12m13n14o15p16q17r18s19t20u21v22w23x24y25z26`. Please see https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens for more information."
  default     = "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"
}

variable "aws_access_key_id" {
  type        = string
  description = "AWS Access Key ID"
}

variable "aws_secret_access_key" {
  type        = string
  description = "AWS Secret Access Key"
}

variable "aws_region" {
  description = "The IAM policy actions to restrict"
  type        = string
  default     = "ap-southeast-1"
}
