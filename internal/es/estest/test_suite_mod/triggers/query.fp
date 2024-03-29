trigger "query" "simple" {
    schedule = "* * * * *"

    enabled = false

    database = "postgres://steampipe:@host.docker.internal:9193/steampipe"
    # database = "mysql://root:flowpipe@tcp(localhost:3306)/flowpipe_test"


    # sql = "select concat(path, '-', cast(key_path as text)) as id, * from config.json_key_value order by id limit 10"
    sql = "select concat(path, '-', cast(key_path as text)) as id, * from config.json_key_value order by id"
    # sql =  "select * from DataTypeDemo"

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

