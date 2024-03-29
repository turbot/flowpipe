pipeline "query" {

    step "query" "query_1" {
        sql      = "select * from aws_account"
        database = "postgres://steampipe:8c6b_44b4_aed9@host.docker.internal:9193/steampipe"
    }

    # step "echo" "result" {
    #     text = "${ join("", [for row in jsondecode(step.query.query_1.rows): "\n- ${row.title}"]) }"
    # }
    step "transform" "result" {
        for_each = step.query.query_1.rows
        value    = each.value.arn
    }
}
