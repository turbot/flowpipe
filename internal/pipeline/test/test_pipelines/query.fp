pipeline "query" {

    step "query" "query_1" {
        sql = "select * from foo"
    }
}

pipeline "query_with_args" {
    step "query" "query_1" {
        sql = "select * from foo where bar = $1 and baz = $2"

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

        args = [
            param.bar,
            param.baz
        ]
    }
}