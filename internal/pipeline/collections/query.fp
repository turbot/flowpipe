pipeline "query" {

    step "query" "query_1" {
        sql = "select * from aws.aws_account"
        connection_string = "postgres://steampipe:8c6b_44b4_aed9@host.docker.internal:9193/steampipe"
    }
}

