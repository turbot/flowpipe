pipeline "pipeline_param_from_var" {

    param "one" {
        type = string
        default = var.var_from_env
    }


    step "echo" "echo" {
        text = param.one
    }

    output "val" {
        value = step.echo.echo
    }
}