pipeline "enum_param_mismatched" {
    param "name" {
        type = list(string)
        enum = [1,2,3]
    }
}