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

# param "iam_policy_restricted_actions" {
#   description = "The IAM policy actions to restrict"
#   type        = string
#   default     = "s3:DeleteBucket,s3:DeleteObject"
# }
