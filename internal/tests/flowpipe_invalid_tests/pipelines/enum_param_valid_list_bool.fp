pipeline "enum_param_valid_lsit_of_bool" {
    param "name" {
        type = list(bool)
        enum = [true, false]
    }
}