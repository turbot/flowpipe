pipeline "enum_param_mismatched" {
    param "name" {
        type = string
        enum = [1,2,3]
    }
}