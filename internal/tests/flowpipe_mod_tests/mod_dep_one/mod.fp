mod "mod_parent" {
  title = "Parent Mod"
  require {
    mod "mod_child_a" {
        version = "1.0.0"
    }
    mod "mod_child_b" {
        version = "1.0.0"
    }
  }
}

# declare trigger before the pipeline to test forward reference
trigger "schedule" "my_hourly_trigger" {
    schedule = "5 * * * *"
    pipeline = pipeline.refer_to_child
}

pipeline "json" {
    step "transform" "json" {
        value = jsonencode({
            Version = "2012-10-17"
            Statement = [
            {
                Action = [
                "ec2:Describe*",
                ]
                Effect   = "Allow"
                Resource = "*"
            },
            ]
        })
    }

    output "foo" {
        value = step.transform.json.value
    }
}

pipeline "refer_to_child" {
    step "pipeline" "child_output" {
        pipeline = mod_child_a.pipeline.this_pipeline_is_in_the_child
    }
}

pipeline "refer_to_child_b" {
    step "pipeline" "child_output" {
        pipeline = mod_child_b.pipeline.foo_two
    }
}


pipeline "foo" {

    # leave this here to ensure that references that is later than the resource can be resolved
    #
    # we parse the HCL files from top to bottom, so putting this step `baz` after `bar` is the easier path
    # reversing is the a harder parse
    step "transform" "baz" {
        value = step.transform.bar
    }

    step "transform" "bar" {
        value = "test"
    }

    step "pipeline" "child_pipeline" {
        pipeline = pipeline.foo_two
    }

    step "transform" "child_pipeline" {
        value = step.pipeline.child_pipeline.foo
    }
}


pipeline "foo_two" {
    step "transform" "baz" {
        value = "foo"
    }

    output "foo" {
        value = transform.baz.value
    }
}
