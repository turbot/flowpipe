pipeline "container_with_loop" {

    step "container" "pipe" {
        image = "alpine:3.7"

        cmd = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]

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
            until = loop.index >= 2
            memory = 150 + loop.index
        }
    }

    output "val" {
        value = step.container.pipe
    }
}

pipeline "container_with_loop_update_cmd" {

    step "container" "pipe" {
        image = "alpine:3.7"

        cmd = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]

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
            until = loop.index >= 2
            memory = 150 + loop.index
            cmd = [ "sh", "-c", "echo '${loop.index}'" ]
        }
    }

    output "val" {
        value = step.container.pipe
    }
}