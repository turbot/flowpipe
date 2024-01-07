pipeline "sqlite_query" {
  step "query" "list" {
    connection_string = "sqlite:./query_source_clean.db"
    sql               = "select * from test_one order by id"
  }

  output "val" {
    value = step.query.list.rows
  }
}

