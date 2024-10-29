pipeline "enum_param_unsupported" {
    param "name" {
        type = map(string)
        enum = ["a", "b", "c"]
    }
}