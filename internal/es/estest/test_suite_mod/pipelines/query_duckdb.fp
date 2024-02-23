pipeline "duckdb_query" {
  step "query" "list" {
    database = "duckdb:./query_duckdb.duckdb"
    sql      = "select * from employee order by id"
  }

  output "val" {
    value = step.query.list.rows
  }
}