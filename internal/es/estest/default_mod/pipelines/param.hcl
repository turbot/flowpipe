pipeline "simple_param" {

    param "name" {
        type = string
        default = "foo"
    }

    step "echo" "name" {
        text = "echo ${param.name}"
    }
}