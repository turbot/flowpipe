pipeline "simple_with_trigger" {
  description = "simple pipeline that will be referred to by a trigger"

  step "transform" "simple_echo" {
    value = "foo bar"
  }
}

trigger "schedule" "my_hourly_trigger" {
  enabled  = false
  schedule = "5 * * * *"
  pipeline = pipeline.simple_with_trigger
}

trigger "schedule" "my_hourly_trigger_interval" {
  enabled  = true
  schedule = "daily"
  pipeline = pipeline.simple_with_trigger
}


trigger "schedule" "trigger_with_args" {
  schedule = "5 * * * *"
  pipeline = pipeline.simple_with_trigger

  args = {
    param_one     = "one"
    param_two_int = 2
  }
}

trigger "http" "trigger_with_args" {
  enabled = true

  method "post" {
    pipeline = pipeline.simple_with_trigger

    args = {
      param_one     = "one"
      param_two_int = 2
    }
  }
}


trigger "query" "query_trigger" {
  schedule = "5 * * * *"

  database = "postgres://steampipe:@host.docker.internal:9193/steampipe"

  sql = <<EOQ
        select
            access_key_id,
            user_name,
            create_date,
            ctx ->> 'connection_name' as connection
        from aws_iam_access_key
        where create_date < now() - interval '90 days'
    EOQ

  primary_key = "access_key_id"
}

trigger "query" "query_trigger_interval" {
  enabled  = true
  schedule = "daily"
  database = "postgres://steampipe:@host.docker.internal:9193/steampipe"

  sql = <<EOQ
        select
            access_key_id,
            user_name,
            create_date,
            ctx ->> 'connection_name' as connection
        from aws_iam_access_key
        where create_date < now() - interval '90 days'
    EOQ

  primary_key = "access_key_id"
}

trigger "http" "trigger_with_execution_mode" {

  method "post" {
    pipeline       = pipeline.simple_with_trigger
    execution_mode = "synchronous"
  }
}
