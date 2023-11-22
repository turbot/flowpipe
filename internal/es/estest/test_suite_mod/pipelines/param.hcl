pipeline "simple_param" {

    param "name" {
        type = string
        default = "if you see this that means something is wrong"
    }

    # test default value coming from var
    param "from_var"  {
        type = string
        default = var.var_one
    }

    step "transform" "name" {
        value = "echo ${param.name}"
    }

    output "val" {
        value = step.transform.name.value
    }

    step "transform" "from_var" {
        value = "echo ${param.from_var}"
    }

    output "from_var" {
        value = step.transform.from_var.value
    }
}


pipeline "calling_pipeline_with_params" {
    step "pipeline" "simple_param" {
        pipeline = pipeline.simple_param

        args = {
            name = "bar"
        }
    }

    step "transform" "foo" {
        value = "foo bar"
    }

    step "pipeline" "simple_param_expr" {
        pipeline = pipeline.simple_param

        args = {
            name = "baz ${step.transform.foo.value}"
        }
    }

    output "val" {
        value = step.pipeline.simple_param.output.val
    }

    output "val_expr" {
        value = step.pipeline.simple_param_expr.output.val
    }

    output "val_from_val" {
        value = step.pipeline.simple_param.output.from_var
    }
}

pipeline "set_param" {

    param "instruments" {
        type = set(string)
        default = ["guitar", "bass", "drums"]
    }

    step "transform" "instruments" {
        for_each = param.instruments
        value    = "[${each.key}] ${each.value}"
    }

    output "val_1" {
        value = step.transform.instruments[0].value
    }
    output "val_2" {
        value = step.transform.instruments[1].value
    }
    output "val_3" {
        value = step.transform.instruments[2].value
    }
    output "val" {
        value = step.transform.instruments
    }
}

pipeline "any_param" {

    param "param_any" {

    }

    step "transform" "echo" {
        value = param.param_any
    }

    output "val" {
        value = step.transform.echo.value
    }
}

pipeline "typed_any_param" {

    param "param_any" {
        type = any
    }

    step "transform" "echo" {
        value = param.param_any
    }

    output "val" {
        value = step.transform.echo.value
    }
}