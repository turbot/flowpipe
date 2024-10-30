pipeline "enum_param_mismatched_2" {
    param "name" {
        type = number
        enum = ["a","b","c"]
    }
}