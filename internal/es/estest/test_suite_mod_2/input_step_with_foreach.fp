
pipeline "root_parent" {

    step "pipeline" "root" {
        pipeline = pipeline.nested_one
    }
}

pipeline "nested_one" {

    step "pipeline" "nested_one" {
        pipeline = pipeline.nested_two
    }

    output "val" {
        value = step.pipeline.nested_one
    }
}

pipeline "nested_two" {

    step "pipeline" "nested_two" {
        for_each = ["one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"]

        pipeline = pipeline.nested_three

        max_concurrency = 1

        args = {
            nested_three_param = each.value
        }
    }

    output "val" {
        value = step.pipeline.nested_two
    }
}

pipeline "nested_three" {
    param "nested_three_param" {
        type = string
    }
    step "pipeline" "nested_three" {
        pipeline = pipeline.nested_four

        args = {
            nested_four_param = param.nested_three_param
        }
    }
}

pipeline "nested_four" {
    param "nested_four_param" {
        type = string
    }

    step "input" "approval" {
        type   = "button"
        prompt = "nested_four ${param.nested_four_param}: Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }
 }