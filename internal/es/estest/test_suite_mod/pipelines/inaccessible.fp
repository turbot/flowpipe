pipeline "bad_http_not_ignored" {
    description = "Pipeline with a HTTP step that will fail. Error is not ignored."

    step "http" "my_step_1" {
        url = "http://google.com/astros.jsons"
    }

    step "transform" "bad_http" {
        depends_on = [step.http.my_step_1]
        value      = "foo"
    }
}


pipeline "inaccessible_fail" {

    step "http" "will_fail" {
        url = "http://api.google.com/bad.json"
        error {
            ignore = true
        }
    }

    # this step will fail because  step.http.will_fail.value does not exist
    step "transform" "will_not_run" {
        value = step.http.will_fail.value
    }
}

pipeline "inaccessible_ok" {

    step "http" "will_fail" {
        url = "http://api.google.com/bad.json"
        error {
            ignore = true
        }
    }

    # this step will NOT fail because  step.http.will_fail does exist just doesn't have a "value"
    step "transform" "will_not_run" {
        value = step.http.will_fail
    }
}