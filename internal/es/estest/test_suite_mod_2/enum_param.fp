pipeline "enum_param" {

    param "string_param" {
        type = string
        default = "value1"
        enum = ["value1", "value2", "value3"]
        tags = {
            "tag3" = "value3"
            "tag4" = "value4"
        }
    }

    param "num_param" {
        type = number
        default = 1
        enum = [1, 2, 3]
        tags = {
            "tag5" = "value5"
            "tag6" = "value6"
        }
    }

    param "list_of_string_param" {
        type = list(string)
        default = ["value1", "value2"]
        enum = ["value1", "value2", "value3"]
        tags = {
            "tag7" = "value7"
            "tag8" = "value8"
        }
    }

    param "aws_conn" {
        type = connection.aws
        default = connection.aws.example
    }

    param "aws_conn_list" {
        type = list(connection.aws)
        default = [connection.aws.example, connection.aws.example_2]

    }

    step "transform" "echo" {
        value = "${param.string_param}"
    }

    step "transform" "echo_2" {
        value = "${param.num_param}"
    }

    step "transform" "echo_3" {
        value = "${param.list_of_string_param}"
    }
}

pipeline "conn_param" {

    param "aws_conn" {
        type = connection.aws
        default = connection.aws.example
    }

    param "generic_conn" {
        type = connection
        default = connection.aws.example
    }

    param "aws_conn_list" {
        type = list(connection.aws)
        default = [connection.aws.example, connection.aws.example_2]

    }

    step "transform" "echo" {
        value = param.aws_conn.access_key
    }

    step "transform" "echo_2" {
        value = param.generic_conn
    }

    output "val" {
        value = step.transform.echo.value
    }

    output "val2" {
        value = step.transform.echo_2.value
    }
}