pipeline "pipeline_param_from_var" {

    param "one" {
        type = string
        default = var.var_from_env
    }


    step "transform" "echo" {
        value = param.one
    }

    output "val" {
        value = step.transform.echo
    }
}