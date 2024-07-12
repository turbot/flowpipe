trigger "query" "simple" {
    schedule = "* * * * *"

    enabled = true

    database = "sqlite:./query_source_clean.db"
    sql      = "select * from test_one order by id"

    primary_key = "id"

    capture "insert" {
        pipeline = pipeline.query_trigger_display
        args = {
            inserted_rows = self.inserted_rows
        }
    }

    capture "update" {
        pipeline = pipeline.query_trigger_display
        args = {
            updated_rows = self.updated_rows
            deleted_rows = self.deleted_rows
        }
    }

    capture "delete" {
        pipeline = pipeline.query_trigger_display
        args = {
            deleted_rows = self.deleted_rows
        }
    }
}


pipeline "query_trigger_display" {

    param "inserted_rows" {
    }

    param "updated_rows" {
    }

    param "deleted_rows" {
    }

    output "inserted_rows" {
        value = param.inserted_rows
    }

    output "updated_rows" {
        value = param.updated_rows
    }

    output "deleted_rows" {
        value = param.deleted_rows
    }
}





pipeline "sqlite_query" {
  step "query" "list" {
    database = "sqlite:./query_source_clean.db"
    sql      = "select * from test_one order by id"
  }

  output "val" {
    value = step.query.list.rows
  }
}

