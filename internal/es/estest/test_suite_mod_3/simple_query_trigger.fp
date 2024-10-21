trigger "query" "simple_sqlite" {
    schedule = "* * * * *"

    enabled = false

    database = "sqlite:./query_source_modified.db"
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
        }
    }

    capture "delete" {
        pipeline = pipeline.query_trigger_display
        args = {
            deleted_rows = self.deleted_rows
        }
    }
}

trigger "query" "simple_sqlite_no_db" {
    schedule = "* * * * *"

    enabled = false

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
        }
    }

    capture "delete" {
        pipeline = pipeline.query_trigger_display
        args = {
            deleted_rows = self.deleted_rows
        }
    }
}

trigger "query" "simple_sqlite_connection" {
    schedule = "* * * * *"

    enabled = false

    database = connection.sqlite.query_source_modified
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
        }
    }

    capture "delete" {
        pipeline = pipeline.query_trigger_display
        args = {
            deleted_rows = self.deleted_rows
        }
    }
}



trigger "query" "simple_sqlite_connection_in_param" {
    schedule = "* * * * *"

    enabled = false

    param "db" {
        type = connection
        default = connection.sqlite.query_source_modified
    }
    database = param.db

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
        }
    }

    capture "delete" {
        pipeline = pipeline.query_trigger_display
        args = {
            deleted_rows = self.deleted_rows
        }
    }
}


