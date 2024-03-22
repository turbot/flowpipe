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