pipeline "bad_http_not_ignored" {
    description = "Pipeline with a HTTP step that will fail. Error is not ignored."
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.jsons"
    }

    step "transform" "bad_http" {
        depends_on = [step.http.my_step_1]
        value      = "foo"
    }
}


pipeline "inaccessible" {

    step "http" "will_fail" {
        url = "http://api.google.com/bad.json"
        error {
            ignore = true
        }
    }

    step "transform" "will_not_run" {
        value = step.http.will_fail
    }
}