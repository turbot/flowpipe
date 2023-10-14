pipeline "run_me" {

    param "name" {
        type = "string"
        default = "value"
    }

    step "echo" "echo" {
        text = "Hello: ${param.name}"
    }

    output "val" {
        value = step.echo.echo.text
    }
}

pipeline "run_me_controller" {

    param "names" {
        type = list(string)
        default = ["spock"]
    }

    step "pipeline" "run" {
        for_each = param.names
        pipeline = pipeline.run_me

        args = {
            name = each.value
        }
    }
}
