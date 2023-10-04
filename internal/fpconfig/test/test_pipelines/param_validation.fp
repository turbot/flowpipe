pipeline "validate_my_param" {

    param "my_token" {
        type = string
    }

    param "my_number" {
        type = number
    }

    param "my_number_two" {
        type = number
    }

    param "my_bool" {
        type = bool
    }

    param "set_string" {
        type = set(string)
    }

    param "set_number" {
        type = set(number)
    }

    param "set_bool" {
        type = set(bool)
    }

    param "set_any" {
        type = set
    }

    param "list_string" {
        type = list(string)
    }

    param "list_bool" {
        type = list(bool)
    }

    param "list_number" {
        type = list(number)
    }

    param "list_number_two" {
        type = list(number)
    }

    param "list_number_three" {
        type = list(number)
    }

    param "list_any" {
        type = list
    }

    param "list_any_two" {
        type = list(any)
    }

    param "list_any_three" {
        type = list
    }

    param "map_of_string" {
        type = map(string)
    }

    param "map_of_number" {
        type = map(number)
    }

    param "map_of_bool" {
        type = map(bool)
    }

    param "map_of_number_two" {
        type = map(number)
    }

    param "map_of_any" {
        type = map
    }

    param "map_of_any_two" {
        type = map
    }

    param "map_of_any_three" {
        type = map
    }

    step "echo" "echo" {
        text = param.my_token
    }
}