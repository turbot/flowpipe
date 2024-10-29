connection "aws" {
  plugin = "aws"

  profile = "silverwater"

  #regions = ["*"] # All regions
  #regions = ["eu-*"] # All EU regions
  #regions = ["us-east-1", "eu-west-2"] # Specific regions

  #default_region = "eu-west-2"

  #profile = "myprofile"

  #max_error_retry_attempts = 9

  #min_error_retry_delay = 25

  #ignore_error_codes = ["AccessDenied", "AccessDeniedException", "NotAuthorized", "UnauthorizedOperation", "UnrecognizedClientException", "AuthorizationError"]

  #endpoint_url = "http://localhost:4566"

  #s3_force_path_style = false
}


connection "aws_keys1" {
  plugin = "aws"
  
  access_key = "abc"
  secret_key = "123"
}
