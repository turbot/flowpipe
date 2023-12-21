trigger "query" "simple" {
    schedule = "* * * * *"

    connection_string = "sqlite:./query_source.db"
    sql = "select * from test_one"
    primary_key = "id"

    pipeline = pipeline.query_trigger_display

    args = {
        inserted_rows = self.inserted_rows
        updated_rows = self.updated_rows
        deleted_keys = self.deleted_keys
    }
}


pipeline "query_trigger_display" {

    param "inserted_rows" {
    }

    param "updated_rows" {
    }

    param "deleted_keys" {
    }

    output "inserted_rows" {
        value = param.inserted_rows
    }

    output "updated_rows" {
        value = param.updated_rows
    }

    output "deleted_keys" {
        value = param.deleted_keys
    }
}

