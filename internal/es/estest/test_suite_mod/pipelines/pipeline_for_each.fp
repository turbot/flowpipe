pipeline "run_me" {

    param "name" {
        type = string
        default = "value"
    }

    step "transform" "echo" {
        value = "Hello: ${param.name}"
    }

    output "val" {
        value = step.transform.echo.value
    }
}

pipeline "run_me_controller" {

    param "names" {
        type = list(string)
        default = ["spock", "kirk", "sulu"]
    }

    step "pipeline" "run" {
        for_each = param.names
        pipeline = pipeline.run_me

        args = {
            name = each.value
        }
    }

    output "val" {
        value = step.pipeline.run
    }
}
