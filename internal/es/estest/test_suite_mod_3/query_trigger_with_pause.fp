trigger "query" "with_pause" {
    schedule = "* * * * *"

    enabled = false

    database = "sqlite:./query_source_modified.db"
    sql      = "select * from test_one order by id"

    primary_key = "id"

    capture "insert" {
        pipeline = pipeline.with_pause
        args = {
            inserted_rows = self.inserted_rows
        }
    }

    capture "update" {
        pipeline = pipeline.with_pause
        args = {
            updated_rows = self.updated_rows
        }
    }

    capture "delete" {
        pipeline = pipeline.with_pause
        args = {
            deleted_rows = self.deleted_rows
        }
    }
}

pipeline "with_pause" {
    param "inserted_rows" {
    }

    param "updated_rows" {
    }

    param "deleted_rows" {
    }

    step "input" "my_step" {
        type   = "button"
        prompt = "Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }

    step "transform" "do_the_thing" {
        depends_on = [step.input.my_step]
        value = step.input.my_step.value
    }

    output "val" {
        value = step.transform.do_the_thing
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