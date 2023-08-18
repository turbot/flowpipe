
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
      aws_s3_bucket
    EOT
}

control "lambda_untagged" {
  title = "Lambda Untagged"
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
      aws_lambda_function
    order by reason
    EOT
}

benchmark "untagged" {
  title = "Untagged"
  children = [
    control.lambda_untagged,
    control.s3_untagged,
  ]
}