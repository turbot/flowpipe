pipeline "run_me_bottom" {

    param "name" {
        type = string
        default = "value bottom"
    }

    step "transform" "echo" {
        value = "bottom: ${param.name}"
    }

    output "val" {
        value = step.transform.echo.value
    }
}

pipeline "run_me_middle" {
    param "names" {
        type = list(string)
        default = ["aaa", "bbb", "ccc"]
    }

    param "name" {
        type = string
        default = "value middle"
    }

    step "pipeline" "run" {
        for_each = param.names

        pipeline = pipeline.run_me_bottom

        args = {
            name = "${each.value} - ${param.name}"
        }
    }

    output "val" {
        value = step.pipeline.run
    }
}

pipeline "run_me_top" {

    param "names" {
        type = list(string)
        default = ["spock", "kirk", "sulu"]
        #default = ["spock"]
    }

    step "pipeline" "run" {
        for_each = param.names
        pipeline = pipeline.run_me_middle

        args = {
            name = each.value
        }
    }

    output "val" {
        value = step.pipeline.run
    }
}
