pipeline "simple_param" {

    param "name" {
        type    = string
        default = "foo"
    }

    step "transform" "name" {
        value = "echo ${param.name}"
    }
}