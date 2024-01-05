pipeline "simple_with_trigger" {
    description = "simple pipeline that will be referred to by a trigger"

    param "param_one" {
        type = string
        default = "this is the default value"
    }

    step "transform" "simple_echo" {
        value = "foo bar: ${param.param_one}"
    }
}

trigger "schedule" "my_every_minute_trigger" {
    schedule = "* * * * *"
    pipeline = pipeline.simple_with_trigger
    args = {
        param_one = "from trigger"
    }
}

trigger "schedule" "my_every_5_minute_trigger" {
    schedule = "*/5 * * * *"
    pipeline = pipeline.simple_with_trigger
}

trigger "schedule" "every_hour_trigger" {
    schedule = "hourly"
    pipeline = pipeline.simple_with_trigger
}