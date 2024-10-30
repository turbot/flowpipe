mod "custom_type_two" {

}

pipeline "custom_type_two" {
    param "conn" {
        default = connection.aws.example
        type = connection.aws
    }

    param "list_of_conns" {
        
        default = [
            connection.aws.example,
            connection.aws.example_2,
            connection.aws.example_3
        ]
        type = list(connection.aws)
    }

    param "conn_generic" {
        type = connection
    }

    param "list_of_conns_generic" {
        type = list(connection)
    }

    param "notifier" {
        type = notifier
    }

    param "list_of_notifier" {
        type = list(notifier)        
    }
}
