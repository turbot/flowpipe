pipeline "loop_sleep" {

    step "sleep" "sleep" {
        duration = "1s"

        loop {
            until = loop.index > 2
            duration = "${loop.index}s"
        }
    }
}

pipeline "loop_http" {

    step "http" "http" {
        url = "http://localhost:7104/loop_http"

        request_body = "initial"

        loop {
            until = loop.index > 2
            request_body = "${result.response_body} - ${loop.index}"
        }
    }
}


pipeline "loop_transform" {

    step "transform" "transform" {
        value = "initial value"

        loop {
            until = loop.index > 2
            value = "${result.value} - ${loop.index}"
        }
    }
}

pipeline "loop_transform_map" {

    step "transform" "transform" {
        value = {
            name = "victor"
            age = 30
        }

        loop {
            until = loop.index > 2
            value = {
                name = "${result.value.name} - ${loop.index}"
                age = result.value.age + loop.index
            }
        }
    }
}


pipeline "loop_container" {

  step "container" "container" {
    image = "alpine:3.7"

    cmd = [ "sh", "-c", "echo -n $FOO" ]

    env = {
      FOO = "bar"
    }

    timeout            = 60000 // in ms
    memory             = 128
    memory_reservation = 64
    memory_swap        = 256
    memory_swappiness  = 10
    read_only          = false
    user               = "root"


    loop {
        until = loop.index > 2
        memory = 128 + loop.index
        env = {
          FOO = "bar - ${loop.index}"
        }
    }
  }
}

pipeline "loop_message" {

    step "message" "message" {
        notifier = notifier.default
        text = "foo"

        loop {
            until = loop.index > 2
            text = "${loop.index}"
        }
    }

    output "val" {
        value = step.message.message
    }
}


pipeline "loop_message_2" {

    step "message" "message" {
        notifier = notifier.default
        text = "foo"

        loop {
            until = loop.index > 2
            text = "${loop.index}"
            notifier = notifier["notifier_${loop.index}"]
        }
    }

    output "val" {
        value = step.message.message
    }
}

pipeline "loop_message_failed" {

    step "message" "message" {
        notifier = notifier.default
        text = "foo"

        loop {
            until = loop.index > 2
            text = "${loop.index}"
            notifier = notifier["notifiers_${loop.index}"]
        }
    }

    output "val" {
        value = step.message.message
    }
}

