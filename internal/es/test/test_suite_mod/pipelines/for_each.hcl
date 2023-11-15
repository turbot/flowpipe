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
