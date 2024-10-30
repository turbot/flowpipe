pipeline "simple_param_with_list_of_string" {

    param "name" {
        type = set(bool)
        default = 23
    }

    step "transform" "name" {
        value = param.name
    }
}