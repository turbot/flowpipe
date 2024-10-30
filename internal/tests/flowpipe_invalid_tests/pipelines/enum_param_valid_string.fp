pipeline "enum_param_valid_string" {
    param "name" {
        type = string
        enum = ["a", "b", "c"]
    }
}