pipeline "enum_param_mismatched_default" {
    param "name" {
        default = "foo"
        enum = [1,2,3]
    }
}