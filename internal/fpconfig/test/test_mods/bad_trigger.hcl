mod "bad_trigger_mod" {
  title = "my_mod"
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

    # if there's no mod.sp
    # http://localhost:7103/api/v0/pipeline/foo/cmd

    # <mod_name>.pipeline.<pipeline_name>
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
}

