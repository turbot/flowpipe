pipeline "enum_param_default_not_in_enum" {
    param "name" {
        type = string
        default = "foo"
        enum = ["bar", "baz"]
    }
}