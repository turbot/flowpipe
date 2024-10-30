mod "custom_type" {

}

pipeline "subtype" {
    param "conn" {
        type    = connection.aws
        default = connection.aws.example
    }

    param "conn_generic" {
        type = connection
        default = connection.aws.example
    }

    param "list_of_conns" {
        type = list(connection.aws)
        default = [
            connection.aws.default,
            connection.aws.example
        ]
    }

    param "list_of_generic_conns" {
        type = list(connection)
        default = [
            connection.aws.default,
            connection.aws.example
        ]
    }
}

