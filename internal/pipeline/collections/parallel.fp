pipeline "two_text" {
    step "text" "text_1" {
        text = "foo"
    }

    step "text" "text_2" {
        text = "baz ${step.text.text_1.text}"
    }
}

pipeline "three_text" {
    step "text" "text_1" {
        text = "foo"
    }

    step "text" "text_2" {
        text = "baz ${step.text.text_1.text}"
    }

    step "text" "text_3" {
        text = "text_1: ${step.text.text_1.text} text_2: ${step.text.text_2.text}"
    }
}

pipeline "http_depends" {
    step "text" "text_1" {
        text = "astros.json"
    }

    step "http" "http_1" {
        url = "http://api.open-notify.org/${step.text.text_1.text}"
    }
}

pipeline "sleep_depends" {
    step "text" "text_1" {
        text = "1s"
    }

    step "sleep" "sleep_1" {
        duration = step.text.text_1.text
    }
}



pipeline "http_and_sleep" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        duration = "2s"
    }
}

pipeline "http_and_sleep_depends" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        depends_on = [step.http.http_1]
        duration = "2s"
    }
}

pipeline "http_and_sleep_multiple_depends" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        depends_on = [step.http.http_1]
        duration = "2s"
    }

    step "http" "http_2" {
        url = "http://api.open-notify.org/astros.json"
        depends_on = [
            step.http.http_1,
            step.sleep.sleep_1
        ]
    }

    step "http" "http_3" {
        url = "http://api.open-notify.org/astros.json"
        depends_on = [
            step.http.http_1,
            step.sleep.sleep_1,
            step.http.http_2
        ]
    }
}


pipeline "two_sleeps" {
    step "sleep" "sleep_1" {
        duration = "1s"
    }

    step "sleep" "sleep_2" {
        duration = "1s"
    }
}