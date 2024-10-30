mod "test_mod" {
  title = "my_mod"

  require {
    flowpipe {
      min_version = "0.1.0"
    }
  }

  tags = {
    foo = "bar"
    green = "day"
  }
}

trigger "schedule" "my_hourly_trigger" {
  schedule = "5 * * * *"
  # this is valid, but not necessary
  # pipeline = local.pipeline.simple_with_trigger
  # this is the recommended way to refer to another resource within the same mod
  pipeline = pipeline.simple_with_trigger

  # this is the way to refer to a pipeline in another mod
  # pipeline = another_mod.pipeline.another_pipeline

  # you can't refer to nested mods
  # pipeline = another_mod.that_other_mod_dependencies.pipeline.that_pipeline

  # http://localhost:7103/api/v0/pipeline/local.pipeline.simple_with_trigger/cmd

  # should this work if simple_with_trigger is the top level mod?
  # http://localhost:7103/api/v0/pipeline/simple_with_trigger/cmd

  # if there's no mod.fp
  # http://localhost:7103/api/v0/pipeline/foo/cmd

  # <mod_name>.pipeline.<pipeline_name>
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
}

pipeline "json_for" {
  step "transform" "json" {
    value = jsonencode({
      Version = "2012-10-17"
      Users   = ["jeff", "jerry", "jim"]
    })
  }


  step "transform" "json_for" {
    for_each = step.transform.json.Users
    value    = "user: ${each.value}"
  }
}

pipeline "simple_with_trigger" {
  description = "simple pipeline that will be referred to by a trigger"

  step "transform" "simple_echo" {
    value = "foo bar"
  }
}

pipeline "inline_documentation" {
  description = "inline doc"
  documentation = "inline pipeline documentation"
}

pipeline "doc_from_file" {
  description = "inline doc"
  documentation = file("./docs/test.md")
}

pipeline "step_with_if_and_depends" {

  step "transform" "one" {
    value = "one"
  }

  step "transform" "two" {
    value = "two"
  }

  step "transform" "three" {
    depends_on = [step.transform.one, step.transform.two]
    
    if = step.transform.one.value == "one"
    value = "${step.transform.one.value} ${step.transform.two.value}"
  }
}