pipeline "query" {

    step "query" "query_1" {
        sql = "select * from foo"
        connection_string = "this is a connection string"
    }
}

pipeline "query_with_args" {
    step "query" "query_1" {
        sql = "select * from foo where bar = $1 and baz = $2"
        connection_string = "this is a connection string"
        args = [
            "two",
            10
        ]
    }
}

pipeline "query_with_args_expr" {
    param "bar" {
        default = "one"
    }

    param "baz" {
        default = 2
    }

    step "query" "query_1" {
        sql = "select * from foo where bar = $1 and baz = $2"
        connection_string = "this is a connection string"

        args = [
            param.bar,
            param.baz
        ]
    }
}