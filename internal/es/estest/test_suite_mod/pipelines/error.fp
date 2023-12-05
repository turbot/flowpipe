pipeline "simple_error" {
    step "http" "does_not_exist" {
        url = "https://google.com/bad.json"
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
