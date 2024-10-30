
credential "aws" "aws_static" {
    access_key = "ASIAQGDFAKEKGUI5MCEU"
    secret_key = "QhLNLGM5MBkXiZm2k2tfake+TduEaCkCdpCSLl6U"
}

pipeline "ex1" {
  param "cred" {
    type    = string
    default = "aws_static"
  }

  step "container" "aws" {
    image = "public.ecr.aws/aws-cli/aws-cli"
    cmd   = [ "s3", "ls" ]
    env   = credential.aws[param.cred].env
  } 
}