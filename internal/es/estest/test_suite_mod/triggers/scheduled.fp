trigger "schedule" "every_minute" {
    enabled = false
    schedule = "* * * * *"
    pipeline = pipeline.two_params

    args = {
        param_one = "joe"
    }
}


pipeline "two_params" {
    param "param_one" {
        default = "default value for param one"
    }

    param "param_two" {
        default = "default value for param two"
    }

    output "one" {
        value = param.param_one
    }

    output "two" {
        value = param.param_two
    }
}