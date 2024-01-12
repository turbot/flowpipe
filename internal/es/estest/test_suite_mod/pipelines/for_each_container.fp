pipeline "parent_with_foreach" {
    param "times" {
        type = number
        default = 3
    }

    step "transform" "names" {
        value = [for i in range(param.times) : "name-${i}"]
    }

    step "pipeline" "nested_one" {
        for_each = step.transform.names.value
        pipeline = pipeline.nested_one
        args = {
            name = each.value
        }
    }

    output "vals" {
        value = step.pipeline.nested_one
    }
}

pipeline "nested_one" {
    param "name" {
        type = string
        default = "default value"
    }
    step "container" "name" {
      image = "bash"
      cmd = ["echo", param.name]
    }
    output "val" {
        value = step.container.name
    }
}