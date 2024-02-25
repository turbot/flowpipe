pipeline "simple_param" {

    param "name" {
        type    = string
        default = "foo"
    }

    step "transform" "name" {
        value = "echo 6 ${param.name}"
    }

    step "transform" "name_2" {
        value = "echo 6 ${param.name}"
    }


    output "val" {
        value = step.transform.name.value
    }
}