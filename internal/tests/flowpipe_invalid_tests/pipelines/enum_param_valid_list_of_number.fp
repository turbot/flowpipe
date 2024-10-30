pipeline "enum_param_valid_string" {
    param "name" {
        type = list(number)
        enum = [1,2,3]
    }
}