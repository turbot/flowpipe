pipeline "simple_error" {
    step "http" "does_not_exist" {
        url = "https://google.com/bad.json"
    }

    output "val" {
        value = "should not be calculated"
    }
}

pipeline "simple_error_ignored_with_if_does_not_match" {
    step "http" "does_not_exist" {
        url = "https://google.com/bad.json"

        error {
            if = result.status_code == 700
            ignore = true
        }
    }

    output "val" {
        value = "should not be calculated"
    }
}

pipeline "simple_error_ignored" {
    step "http" "does_not_exist" {
        url = "https://google.com/bad.json"

        error {
            ignore = true
        }
    }

    output "val" {
        value = "should be calculated"
    }
}

pipeline "simple_error_ignored_multi_steps" {
    step "http" "does_not_exist" {
        url = "https://google.com/bad.json"

        error {
            ignore = true
        }
    }

    step "transform" "two" {
        depends_on = [step.http.does_not_exist]
        value = "should exist"
    }

    output "val" {
        value = "should be calculated"
    }

    output "val_two" {
        value = step.transform.two.value
    }
}



pipeline "failed_output_calc" {
    step "transform" "echo" {
        value = "echo that works"
    }

    output "val_error" {
        value = step.transform.echo.bar
    }

    output "val_error_two" {
        value = step.transform.echo.bar
    }

    output "val_ok" {
        value = step.transform.echo.value
    }

    output "val_ok_two" {
        value = "this works"
    }
}

pipeline "parent_with_child_with_no_output" {
    step "pipeline" "call_child" {
        pipeline = pipeline.child_with_no_output
    }

    output "val" {
    value       = {
      "call_child" = !is_error(step.pipeline.call_child) ? "ok" : "fail"
    }
  }
}

pipeline "child_with_no_output" {

    step "transform" "echo" {
        value = "echo"
    }
}

pipeline "step_output_should_not_calculate_if_error" {
    step "http" "bad" {
        url = "https://google.com/abc.json"

        output "val" {
            value = "step: should not be calculated"
        }
    }

    output "val" {
        value = "pipeline: should not be calculated"
    }
}

pipeline "step_output_should_be_calculated_because_step_error_is_ignored" {
    step "http" "bad" {
        url = "https://google.com/abc.json"

        error {
            ignore = true
        }

        output "val" {
            value = "step: should be calculated"
        }
    }

    output "val" {
        value = "pipeline: should be calculated"
    }
}

