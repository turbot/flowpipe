trigger "query" "my_query_trigger" {
  database    = "postgres://steampipe@localhost:9193/steampipe"
  primary_key = "arn"

  sql = <<EOQ
      select
        arn,
        instance_id,
        instance_state,
        instance_type,
        region,
        account_id
      from
        aws_ec2_instance;
  EOQ

  capture "insert" {
    pipeline = pipeline.instance_added

    args = {
      rows = self.inserted_rows
    }
  }

  capture "update" {
    pipeline = pipeline.instance_changed

    args = {
      rows = self.updated_rows
    }
  }

  capture "delete" {
    pipeline = pipeline.instance_terminated

    args = {
      rows = self.deleted_rows
    }
  }
}

pipeline "instance_added" {
  param "rows" {
  }

  step "transform" "echo" {
    value = param.rows
  }

  output "val" {
    value = step.transform.echo.value
  }
}

pipeline "instance_changed" {
  param "rows" {
  }

  step "transform" "echo" {
    value = param.rows
  }

  output "val" {
    value = step.transform.echo.value
  }
}

pipeline "instance_terminated" {
  param "rows" {
  }

  step "transform" "echo" {
    value = param.rows
  }

  output "val" {
    value = step.transform.echo.value
  }
}


