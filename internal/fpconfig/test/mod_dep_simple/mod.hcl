mod "mod_parent" {
  title = "Parent Mod"
  require {
    mod "mod_child_a" {
        version = "1.0.0"
        args = {
            var_two = var.var_two_parent,
        }
    }
  }
}


variable "var_two_parent" {
  type        = string
  description = "test variable"
}

pipeline "json" {
    step "echo" "json" {
        json = jsonencode({
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
        value = step.echo.json.json
    }
}

pipeline "foo" {

    # leave this here to ensure that references that is later than the resource can be resolved
    #
    # we parse the HCL files from top to bottom, so putting this step `baz` after `bar` is the easier path
    # reversing is the a harder parse
    step "echo" "baz" {
        text = step.echo.bar
    }

    step "echo" "bar" {
        text = "test"
    }
}


pipeline "refer_to_child" {
    step "pipeline" "child_output" {
        pipeline = mod_child_a.pipeline.this_pipeline_is_in_the_child
    }
}
