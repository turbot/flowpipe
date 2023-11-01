mod "test_suite_mod" {
    title = "Default test suite mod 3"
    require {
        mod "mod_depend_a" {
            version = "1.0.0"
            args = {
                var_depend_a_one = var.var_one
                var_depend_a_two = var.var_depend_a_two
            }
        }
    }

}

pipeline "echo_one_a" {
    description = "foo"

    step "echo" "echo_one" {
        text = "Hello World"
    }

    step "pipeline" "child_pipeline" {
        pipeline = mod_depend_a.pipeline.echo_one_depend_a
    }

    output "echo_one_output" {
        value = step.pipeline.child_pipeline.output.val
    }

    output "echo_one_output_val_var_one" {
        value = step.pipeline.child_pipeline.output.val_var_one
    }
}


pipeline "echo_one" {
    step "echo" "echo_one" {
        text = "Hello World"
    }

    step "pipeline" "child_pipeline" {
        pipeline = mod_depend_a.pipeline.echo_one_depend_a
    }

    output "echo_one_output" {
        value = step.pipeline.child_pipeline.output.val
    }

    output "echo_one_output_val_var_one" {
        value = step.pipeline.child_pipeline.output.val_var_one
    }
}

pipeline "echo_with_variable" {
    step "echo" "echo_one" {
        text = "Hello World: ${var.var_one}"
    }

    step "echo" "echo_two" {
        text = "Hello World Two: ${var.var_two}"
    }

    step "echo" "echo_three" {
        text = "Hello World Two: ${var.var_two} and ${step.echo.echo_two.text}"
    }

    step "echo" "echo_four" {
        text = local.locals_one
    }

    step "echo" "echo_five" {
        text = "${local.locals_two} AND ${step.echo.echo_two.text} AND ${step.echo.echo_four.text}"
    }

    output "echo_one_output" {
        value = step.echo.echo_one.text
    }

    output "echo_two_output" {
        value = step.echo.echo_two.text
    }

    output "echo_three_output" {
        value = step.echo.echo_three.text
    }

    output "echo_four_output" {
        value = step.echo.echo_four.text
    }

    output "echo_five_output" {
        value = step.echo.echo_five.text
    }
}

pipeline "expr_depend_and_function" {
    step "echo" "text_1" {
        text = "foo bar"
    }

    step "echo" "text_2" {
        text = "lower case ${title("bar ${step.echo.text_1.text} baz")} and here"
    }

    step "echo" "text_3" {
        text = "output 2 ${title(step.echo.text_2.text)} title(output1) ${title(step.echo.text_1.text)}"
    }

    step "echo" "explicit_depends" {
        depends_on = [
            step.echo.text_2,
            step.echo.text_1
        ]
        text = "explicit depends here"
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

    step "echo" "echo_sleep_for" {
        for_each = step.sleep.sleep_1
        text = each.value.duration
    }

    step "echo" "echo_sleep_1" {
        text = "sleep 2 output: ${step.echo.echo_sleep_for[1].text}"
    }

    step "echo" "echo_sleep_2" {
        text = "sleep 1 output: ${step.sleep.sleep_1[0].duration}"
    }

    step "echo" "echo_for_if" {
        for_each = step.sleep.sleep_1
        text = "sleep 1 output: ${each.value.duration}"
        if = each.value.duration == "1s"
    }


    step "echo" "literal_for" {
        for_each = ["bach", "beethoven", "mozart"]
        text = "name is ${each.value}"
    }


    param "user_data" {
        type = map(list(string))
        default = {
            Users = ["shostakovitch", "prokofiev", "rachmaninoff"]
        }
    }

    step "echo" "literal_for_from_list" {
        for_each = { for user in param.user_data.Users : user => user }
        text = each.value
    }

    output "one" {
        value = step.echo.echo_sleep_2.text
    }

    output "one_function" {
        value = title(step.echo.echo_sleep_2.text)
    }

    output "indexed" {
        value = step.echo.echo_sleep_for[1].text
    }
}