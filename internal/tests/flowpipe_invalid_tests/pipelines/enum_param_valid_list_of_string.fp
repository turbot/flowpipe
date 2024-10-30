pipeline "enum_param_valid_string" {
    param "name" {
        type = list(string)
        enum = ["a", "b", "c"]
    }
}