trigger "query" "with_connection" {
    schedule = "* * * * *"

    enabled = false

    database = connection.steampipe.default

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

trigger "query" "with_connection_in_param" {
    schedule = "* * * * *"

    enabled = false

    param "db" {
        type = connection.steampipe
        default = connection.steampipe.default
    }
    
    database = param.db

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