pipeline "sqlite_query" {
  step "query" "list" {
    connection_string = "sqlite:./query_source_clean.db"
    sql               = "select * from test_one order by id"
  }

  output "val" {
    value = step.query.list.rows
  }
}

pipeline "sqlite_query_path_alternate_b" {
  step "query" "list" {
    connection_string = "sqlite://./query_source_clean.db"
    sql               = "select * from test_one order by id"
  }

  output "val" {
    value = step.query.list.rows
  }
}

pipeline "sqlite_query_path_alternate_c" {
  step "query" "list" {
    connection_string = "sqlite://query_source_clean.db"
    sql               = "select * from test_one order by id"
  }

  output "val" {
    value = step.query.list.rows
  }
}


pipeline "sqlite_query_with_timeout" {
  step "query" "list" {
    connection_string = "sqlite:./query_source_clean.db"
    sql               = <<EOT
  with recursive fibo (curr, next) as (
    select 1, 1
	  union all
	  select next, curr+next from fibo limit 10000000
  )
	select group_concat(curr) from fibo;
   EOT
    timeout           = "50ms"
  }

  output "val" {
    value = step.query.list.rows
  }
}
