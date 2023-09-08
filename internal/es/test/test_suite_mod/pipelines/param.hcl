pipeline "simple_param" {

    param "name" {
        type = string
        default = "if you see this that means something is wrong"
    }

    step "echo" "name" {
        text = "echo ${param.name}"
    }

    output "val" {
        value = step.echo.name.text
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
}