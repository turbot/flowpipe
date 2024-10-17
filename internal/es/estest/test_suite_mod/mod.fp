mod "test_suite_mod" {
    title = "Default Test Suite Mod"

    description = "Test Suite Mode for Flowpipe"

    categories = [
        "Test Suite",
        "Flowpipe"
    ]

    opengraph {
        title       = "Flowpipe Test Suite Mod"
        description = "Run pipelines to supercharge your workflows using Flowpipe."
    }

    require {
        mod "mod_depend_a" {
            version = "1.0.0"
            args = {
                var_depend_a_one = var.var_one

                # race condition bug I think caused by this commit: https://github.com/turbot/pipe-fittings/commit/7f5fc0de25d6cb3bc4a4d5dff40b038770ddca2e
                #
                # Error msg:
                # failed to resolve dependency mod argument value: "var.var_depend_a_two = var.var_depend_a_two" (mod.test_suite_mod /home/runner/work/flowpipe/flowpipe/flowpipe/internal/es/estest/test_suite_mod/mod.fp:21)
                #
                # it's intermittent, thus the race condition suspicion
                var_depend_a_two = "abc"
            }
        }

        mod "mod_depend_c" {
            version = "2.0.0"
        }
    }

}

pipeline "echo_one_a" {
    description = "foo"

    step "transform" "echo_one" {
        value = "Hello World"
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
    step "transform" "echo_one" {
        value = "Hello World"
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
    step "transform" "echo_one" {
        value = "Hello World: ${var.var_one}"
    }

    step "transform" "echo_two" {
        value = "Hello World Two: ${var.var_two}"
    }

    step "transform" "echo_three" {
        value = "Hello World Two: ${var.var_two} and ${step.transform.echo_two.value}"
    }

    step "transform" "echo_four" {
        value = local.locals_one
    }

    step "transform" "echo_five" {
        value = "${local.locals_two} AND ${step.transform.echo_two.value} AND ${step.transform.echo_four.value}"
    }

    output "echo_one_output" {
        value = step.transform.echo_one.value
    }

    output "echo_two_output" {
        value = step.transform.echo_two.value
    }

    output "echo_three_output" {
        value = step.transform.echo_three.value
    }

    output "echo_four_output" {
        value = step.transform.echo_four.value
    }

    output "echo_five_output" {
        value = step.transform.echo_five.value
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
        type    = list(string)
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
        type    = map(list(string))
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