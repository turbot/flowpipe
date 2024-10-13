pipeline "foreach_with_conn_object" {

    param "foo" {
        type = string
        default = "bar"
    }
    step "transform" "source" {
        value = ["example", "example_2", "example_3"]
    }

    step "transform" "repeat" {
        for_each = step.transform.source.value
        value = {
            "value" = "bar"
            "param_foo" = param.foo
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