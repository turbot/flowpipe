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

    step "echo" "name" {
        text = "echo ${param.name}"
    }

    output "val" {
        value = step.echo.name.text
    }

    step "echo" "from_var" {
        text = "echo ${param.from_var}"
    }

    output "from_var" {
        value = step.echo.from_var.text
    }
}


pipeline "calling_pipeline_with_params" {
    step "pipeline" "simple_param" {
        pipeline = pipeline.simple_param

        args = {
            name = "bar"
        }
    }

    step "echo" "foo" {
        text = "foo bar"
    }

    step "pipeline" "simple_param_expr" {
        pipeline = pipeline.simple_param

        args = {
            name = "baz ${step.echo.foo.text}"
        }
    }

    output "val" {
        value = step.pipeline.simple_param.val
    }

    output "val_expr" {
        value = step.pipeline.simple_param_expr.val
    }

    output "val_from_val" {
        value = step.pipeline.simple_param.from_var
    }
}

pipeline "set_param" {

    param "instruments" {
        type = set(string)
        default = ["guitar", "bass", "drums"]
    }

    step "echo" "instruments" {
        for_each = param.instruments
        text = "[${each.key}] ${each.value}"
    }

    output "val_1" {
        value = step.echo.instruments[0].text
    }
    output "val_2" {
        value = step.echo.instruments[1].text
    }
    output "val_3" {
        value = step.echo.instruments[2].text
    }
    output "val" {
        value = step.echo.instruments
    }
}