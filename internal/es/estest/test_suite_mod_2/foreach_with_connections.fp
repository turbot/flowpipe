pipeline "foreach_with_conn_object" {

    param "foo" {
        type = string
        default = "example_4"
    }

    param "bar" {
        type = string
        default = "bar"
    }

    step "transform" "source" {
        value = ["example", "example_2", "example_3"]
    }

    step "transform" "repeat" {
        for_each = step.transform.source.value
        value = {
            "obj value" = "bar + ${connection.aws[param.foo].access_key} + ${param.bar}"
            "param_foo" = param.foo
            "param_bar" = param.bar
            "akey" = connection.aws[each.value]
        }
    }

    output "val" {
        value = step.transform.repeat
    }
}


pipeline "foreach_with_conn_literal" {

    param "foo" {
        type = string
        default = "bar"
    }
    step "transform" "source" {
        value = ["example", "example_2", "example_3"]
    }

    step "transform" "repeat" {
        for_each = step.transform.source.value
        value = "Foo: ${param.foo} and ${connection.aws[each.value].access_key}"
    }

    output "val" {
        value = step.transform.repeat
    }
}

pipeline "foreach_with_conn_simple" {

    step "transform" "source" {
        value = ["example", "example_2", "example_3"]
    }

    step "transform" "repeat" {
        for_each = step.transform.source.value
        value = connection.aws[each.value]
    }

    output "val" {
        value = step.transform.repeat
    }
}

pipeline "from_param" {

    param "cred" {
        type = string
        default = "example"
    }

    step "transform" "next" {
        value = connection.aws[param.cred]
    }

    output "val" {
        value = step.transform.next
    }
}

pipeline "from_another_step" {

    step "transform" "source" {
        value = "example_2"
    }

    step "transform" "next" {
        value = connection.aws[step.transform.source.value]
    }

    output "val" {
        value = step.transform.next
    }
}