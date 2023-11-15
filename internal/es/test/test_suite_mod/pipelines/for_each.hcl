pipeline "for_each_empty_test" {

    param "input" {
        type = list(string)
        default = []
    }

    step "echo" "echo" {
        for_each = param.input
        text = each.value
    }


    output "echo" {
        value = step.echo.echo
    }
}

pipeline "for_each_non_collection" {

    param "input" {
        type = string
        default = "foo"
    }

    step "echo" "echo" {
        for_each = param.input
        text = each.value
    }

    output "echo" {
        value = step.echo.echo
    }
}

pipeline "for_each_and_for_each" {

    step "transform" "first" {
        for_each = ["bach", "mozart", "beethoven"]
        value = each.value
    }

    step "transform" "second" {
        depends_on = [step.transform.first]
        for_each = ["coltrane", "davis", "monk"]

        value = each.value
    }

    step "transform" "third" {
        for_each = step.transform.first

        value = "value is: ${each.value.value}"
    }

    output "first" {
        value = step.transform.first
    }

    output "second" {
        value = step.transform.second
    }

    output "third" {
        value = step.transform.third
    }
}

pipeline "for_each_one_and_for_each_two" {

    step "transform" "first" {
        value = "I'm first"
    }

    step "transform" "echo" {
        depends_on = [step.transform.first]

        for_each = ["bar", "baz", "qux", "quux", "corge", "grault", "garply", "waldo", "fred", "plugh", "xyzzy", "thud"]

        value = "${each.key}: foo ${each.value}"

        output "val" {
            value = "val is: ${each.value}"
        }
    }

    step "transform" "last" {
        depends_on = [step.transform.echo]

        value = "I'm last and should only run once"
    }

    output "val" {
        value = step.transform.last
    }
}
