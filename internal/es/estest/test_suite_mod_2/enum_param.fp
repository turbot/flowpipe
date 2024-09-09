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
        type = string
        default = "example"
        subtype = connection.aws
    }

    param "aws_conn_list" {
        type = list(string)
        default = ["example", "example_2"]
        subtype = list(connection.aws)
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