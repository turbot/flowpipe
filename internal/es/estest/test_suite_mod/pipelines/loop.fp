pipeline "simple_loop" {

    step "transform" "repeat" {
        value  = "iteration: ${loop.index}"

        loop {
            until = loop.index >= 2
            value = "override for ${loop.index}"
        }
    }

    output "val_1" {
        value = step.transform.repeat["0"].value
    }
    output "val_2" {
        value = step.transform.repeat["1"].value
    }
    output "val_3" {
        value = step.transform.repeat["2"].value
    }

    output "val_all" {
        value = step.transform.repeat
    }
}

pipeline "simple_loop_index" {

    step "transform" "repeat" {
        value  = "iteration: ${loop.index}"

        loop {
            until = loop.index >= 2
        }
    }

    output "val_1" {
        value = step.transform.repeat["0"].value
    }
    output "val_2" {
        value = step.transform.repeat["1"].value
    }
    output "val_3" {
        value = step.transform.repeat["2"].value
    }
}


pipeline "loop_with_for_each" {

    step "transform" "repeat" {
        for_each = ["oasis", "blur", "radiohead"]
        value = "iteration: ${loop.index} - ${each.value}"

        loop {
            until = loop.index >= 3
        }
    }

    output "val" {
        value = step.transform.repeat
    }
}

pipeline "lots_of_for_each" {

    step "transform" "repeat" {
        for_each = ["oasis", "blur", "radiohead", "the verve", "the beatles", "the rolling stones", "the sex pistols"]
        value = "name: ${each.value}"
    }

    output "val" {
        value = step.transform.repeat
    }
}


pipeline "loop_with_for_each_sleep" {

    step "sleep" "repeat" {
        for_each = ["1s", "4s"]
        duration = each.value

        loop {
            until = loop.index >= 2
        }
    }

    output "val" {
        value = step.sleep.repeat
    }
}

pipeline "loop_with_for_each_and_nested_pipeline" {

    step "pipeline" "repeat" {
        for_each = ["oasis", "blur", "radiohead"]
        # for_each = ["oasis"]
        pipeline = pipeline.nested_echo

        args = {
            name = each.value
            loop_index = loop.index
        }

        loop {
            until = loop.index >= 2
        }
    }

    output "val" {
        value = step.pipeline.repeat
    }
}

pipeline "nested_echo" {

    param "name" {
        type = string
    }

    param "loop_index" {
        type = number
    }

    step "transform" "echo" {
        value = "${param.loop_index}: ${param.name}"
    }

    output "val" {
        value = step.transform.echo.value
    }
}

pipeline "request_body_loop" {

  step "http" "random" {
    url = "http://localhost:7104/special-case"
    method = "post"
    request_body = jsonencode({
      query = "bar"
    })

    loop {
      until = loop.index >= 2
      request_body = replace(result.request_body, "bar", "baz")
    }
  }

  output "val" {
    value = step.http.random
  }
}
