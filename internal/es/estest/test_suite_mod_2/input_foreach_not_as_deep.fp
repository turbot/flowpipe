pipeline "foreach_parent" {

    step "pipeline" "nested_two" {
        for_each = ["one", "two", "three"]

        pipeline = pipeline.foreach_child

        max_concurrency = 1

        args = {
            foreach_child = each.value
        }
    }
}

pipeline "foreach_child" {
    param "foreach_child" {
        type = string
    }

    step "input" "my_step" {
        type   = "button"
        prompt = "${param.foreach_child} - Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }
}