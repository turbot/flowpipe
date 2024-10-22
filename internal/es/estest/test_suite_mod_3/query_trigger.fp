trigger "query" "simple" {
    schedule = "1 * * * 1"

    database = param.database_connection

    enabled = false

    param "database_connection" {
        type = string
        default = "postgres://steampipe:@localhost:9193/steampipe"
    }

    param "sql" {
        type = string
        default = "select * from aws_s3_bucket"
    }

    sql = param.sql

    param "primary_key" {
        type = string
        default = "arn"
    }

    primary_key = param.primary_key


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
