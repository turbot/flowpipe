pipeline "complex_one" {
    description = "Demo #5 - Delete Turbot Pipes snapshots older than max_days"

    param "identity_type" {
        default = "org"
    }

    param "identity" {
        default = "vandelay-industries"
    }

    param "workspace" {
        default = "latex"
    }

    param "max_days" {
        default = 120
    }

    step "http" "run_query" {
        url    = join("/", ["https://cloud.steampipe.io/api/latest",
                            param.identity_type,
                            param.identity,
                            "workspace",
                            param.workspace,
                            "snapshot?where=${urlencode("created_at < now() - interval '${param.max_days} days'")}"])
        method = "get"

        request_headers = {
          Authorization = "Bearer ${file("./demo.fp")}"
          Content-Type  = "application/json"
        }
    }

    step "http" "send_to_slack" {
        for_each = jsondecode(step.http.run_query.response_body).items
        url      = "https://hooks.slack.com/services/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"
        method   = "post"

        request_body   = jsonencode({
            text = "Snapshot \"${each.value.title}\" (${each.value.id}) created at ${each.value.created_at} is older than ${param.max_days} days and will be deleted."
        })
    }

    step "http" "delete_snap" {
        for_each   = jsondecode(step.http.run_query.response_body).items
        depends_on = [step.http.send_to_slack]
        method     = "delete"
        url        = join("/", ["https://cloud.steampipe.io/api/latest",
                            param.identity_type,
                            param.identity,
                            "workspace",
                            param.workspace,
                            "snapshot",
                            "${each.value.id}"])

        request_headers = {
          Authorization = "Bearer ${file("./demo.fp")}"
          Content-Type  = "application/json"
        }

        error {
            ignore = true
        }
    }


    step "http" "send_error_to_slack" {
        for_each = step.http.delete_snap
        if       = is_error(each.value)
        url      = "https://hooks.slack.com/services/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"
        method   = "post"

        request_body   = jsonencode({
            text = "Deletion failed for snapshot: ${each.value.response_body})."
        })
    }

    step "http" "send_success_to_slack" {
        for_each = step.http.delete_snap
        if       = !is_error(each.value)
        url      = "https://hooks.slack.com/services/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"
        method   = "post"

        request_body   = jsonencode({
            text = "Deletion succeeded for snapshot: ${jsondecode(each.value.response_body).id})."
        })
    }

}