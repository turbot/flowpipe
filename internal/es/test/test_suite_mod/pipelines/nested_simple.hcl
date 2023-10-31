pipeline "nested_simple_top" {

    step "pipeline" "middle" {
        pipeline = pipeline.nested_simple_middle
    }


    output "val" {
        value = step.pipeline.middle.output.val
    }

    output "val_two" {
        value = step.pipeline.middle.output.val_two
    }
}

pipeline "nested_simple_middle" {

    param "input" {
        default = "no name band"
    }

    step "echo" "echo" {
        text = "hello from the middle world"
    }

    output "val" {
        value = step.echo.echo.text
    }

    output "val_two" {
        value = "two: ${step.echo.echo.text}"
    }

    output "val_param" {
        value = param.input
    }
}

pipeline "nested_simple_top_with_merged_output" {

    step "pipeline" "middle" {
        pipeline = pipeline.nested_simple_middle

        output "step_output" {
            value = "step output"
        }
    }

    output "val" {
        value = step.pipeline.middle.output.val
    }

    output "val_two" {
        value = step.pipeline.middle.output.val_two
    }

    output "val_step_output" {
        value = step.pipeline.middle.output.step_output
    }
}

pipeline "nested_simple_top_with_for_each" {

    step "pipeline" "middle" {
        for_each = ["hot mulligan", "sugarcult", "the wonder years"]
        # for_each = ["hot mulligan"]
        pipeline = pipeline.nested_simple_middle

        args = {
            input = each.value
        }
    }

    output "val" {
        value = step.pipeline.middle
    }
}


pipeline "nested_simple_top_with_for_each_with_merged_output" {

    step "pipeline" "middle" {
        for_each = ["hot mulligan", "sugarcult", "the wonder years"]

        pipeline = pipeline.nested_simple_middle

        args = {
            input = each.value
        }

        output "step_output" {
            value = "band: ${each.value}"
        }
    }

    output "step_output_1" {
        value = step.pipeline.middle["0"].output.step_output
    }

    output "step_output_2" {
        value = step.pipeline.middle["1"].output.step_output
    }

    output "step_output_3" {
        value = step.pipeline.middle["2"].output.step_output
    }

    output "val_param_1" {
        value = step.pipeline.middle["0"].output.val_param
    }
    output "val_param_2" {
        value = step.pipeline.middle["1"].output.val_param
    }
    output "val_param_3" {
        value = step.pipeline.middle["2"].output.val_param
    }

}