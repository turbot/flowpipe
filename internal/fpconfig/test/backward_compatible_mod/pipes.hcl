pipeline "parent_pipeline_hcl" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo_b" {
        value = step.echo.foo.text
    }
}


pipeline "parent_pipeline_hcl_b" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo_b" {
        value = step.echo.foo.text
    }
}
