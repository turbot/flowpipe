locals {
  ebs_common_tags = merge(local.aws_thrifty_common_tags, {
    service = "AWS/EBS"
  })
}
