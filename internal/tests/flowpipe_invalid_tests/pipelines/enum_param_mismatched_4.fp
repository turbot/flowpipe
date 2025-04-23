pipeline "enum_param_mismatched" {
    param "name" {
        type = bool
        enum = [1,2,3]
    }
}