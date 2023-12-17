trigger "query" "simple" {
    schedule = "* * * * *"
    connection_string = "postgres://steampipe@host.docker.internal:9193/steampipe"
    sql = "select * from hackernews.hackernews_new"

    primary_key = "id"

    pipeline = pipeline.simple_with_trigger
}
