mod "default_mod" {
    title = "Default built-in Mod"
    require {
        mod "mod_depend_a" {
            version = "1.0.0"
        }
    }

}

pipeline "echo_one" {
    step "transform" "echo_one" {
        value = "Hello World"
    }

    step "pipeline" "child_pipeline" {
        pipeline = mod_depend_a.pipeline.echo_one_depend_a
    }

    output "echo_one_output" {
        value = step.pipeline.child_pipeline.output.val
    }
}

pipeline "expr_depend_and_function" {
    step "transform" "text_1" {
        value = "foo bar"
    }

    step "transform" "text_2" {
        value = "lower case ${title("bar ${step.transform.text_1.value} baz")} and here"
    }

    step "transform" "text_3" {
        value = "output 2 ${title(step.transform.text_2.value)} title(output1) ${title(step.transform.text_1.value)}"
    }

    step "transform" "explicit_depends" {
        depends_on = [
            step.transform.text_2,
            step.transform.text_1
        ]
        value = "explicit depends here"
    }

    # "time"/"for"/"sleep" steps
     param "time" {
        type = list(string)
        default = ["1s", "2s"]
    }

    step "sleep" "sleep_1" {
        for_each = param.time
        duration = each.value
    }

    step "transform" "echo_sleep_for" {
        for_each = step.sleep.sleep_1
        value    = each.value.duration
    }

    step "transform" "echo_sleep_1" {
        value = "sleep 2 output: ${step.transform.echo_sleep_for[1].value}"
    }

    step "transform" "echo_sleep_2" {
        value = "sleep 1 output: ${step.sleep.sleep_1[0].duration}"
    }

    step "transform" "echo_for_if" {
        for_each = step.sleep.sleep_1
        value    = "sleep 1 output: ${each.value.duration}"
        if       = each.value.duration == "1s"
    }


    step "transform" "literal_for" {
        for_each = ["bach", "beethoven", "mozart"]
        value    = "name is ${each.value}"
    }


    param "user_data" {
        type = map(list(string))
        default = {
            Users = ["shostakovitch", "prokofiev", "rachmaninoff"]
        }
    }

    step "transform" "literal_for_from_list" {
        for_each = { for user in param.user_data.Users : user => user }
        value    = each.value
    }

    output "one" {
        value = step.transform.echo_sleep_2.value
    }

    output "one_function" {
        value = title(step.transform.echo_sleep_2.value)
    }

    output "indexed" {
        value = step.transform.echo_sleep_for[1].value
    }
}