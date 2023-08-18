
mod "untagged_example" {
  title = "Untagged Examples"
}

control "s3_untagged" {
  title = "S3 Untagged"
  sql = <<EOT
    select
      arn as resource,
      case
        when tags is not null then 'ok'
        else 'alarm'
      end as status,
      case
        when tags is not null then name || ' has tags.'
        else name || ' has no tags.'
      end as reason,
      region,
      account_id
    from
      aws.aws_s3_bucket
    where region = 'ap-southeast-1'
    EOT
}

control "s3_untagged_b" {
  title = "S3 Untagged"
  sql = <<EOT
    select
      arn as resource,
      case
        when tags is not null then 'ok'
        else 'alarm'
      end as status,
      case
        when tags is not null then name || ' has tags.'
        else name || ' has no tags.'
      end as reason,
      region,
      account_id
    from
      aws.aws_s3_bucket
    where region = 'ap-southeast-2'
    EOT
}

control "s3_untagged_c" {
  title = "S3 Untagged"
  sql = <<EOT
    select
      arn as resource,
      case
        when tags is not null then 'ok'
        else 'alarm'
      end as status,
      case
        when tags is not null then name || ' has tags.'
        else name || ' has no tags.'
      end as reason,
      region,
      account_id
    from
      aws.aws_s3_bucket
    where region = 'ap-south-1'
    EOT
}



benchmark "untagged" {
  title = "Untagged"
  children = [
    control.s3_untagged,
    control.s3_untagged_b,
  ]
}