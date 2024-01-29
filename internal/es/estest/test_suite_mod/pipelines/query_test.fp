pipeline "query_local_test" {
  step "query" "list" {
    connection_string = "postgres://steampipe:xxxxx@host.docker.internal:9193/steampipe"
    sql               = "select concat(path, '-', cast(key_path as text)) as id, path, key_path, keys from config.json_key_value order by id limit 10;"
  }

  output "val" {
    value = step.query.list.rows
  }
}
