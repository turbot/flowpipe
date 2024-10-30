pipeline "simple_param_with_list_of_string" {

    param "name" {
        type = list(string)
        default = "if you see this that means something is wrong"
    }

    step "transform" "name" {
        value = param.name
    }

    output "val" {
        value = step.transform.name.value
    }
}