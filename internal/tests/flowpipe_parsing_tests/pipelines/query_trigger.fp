
pipeline "simple_with_trigger" {
  description = "simple pipeline that will be referred to by a trigger"

  step "transform" "simple_echo" {
    value = "foo bar"
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

  capture "insert" {
    pipeline = pipeline.simple_with_trigger
    args = {
      rows = self.inserted_rows
    }
  }

  capture "update" {
    pipeline = pipeline.simple_with_trigger
    args = {
      rows = self.updated_rows
    }
  }

  capture "delete" {
    pipeline = pipeline.simple_with_trigger
    args = {
      rows = self.deleted_rows
    }
  }
}


// No schedule = every 15 minutes
trigger "query" "query_trigger_no_schedule" {
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
}
