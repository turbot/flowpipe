pipeline "simple_container_step" {

  step "container" "container_test_1" {
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
  }

  output "stdout" {
    value = step.container.container_test_1.stdout
  }

  output "stderr" {
    value = step.container.container_test_1.stderr
  }

  output "exit_code" {
    value = step.container.container_test_1.exit_code
  }

  output "lines" {
    value = step.container.container_test_1.lines[*]
  }
}

pipeline "simple_container_step_with_param" {

  param "image" {
    type    = string
    default = "alpine:3.7"
  }

  param "cmd" {
    type    = list(string)
    default = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]
  }

  param "env" {
    type    = map(string)
    default = {
      FOO = "bar"
    }
  }

  param "timeout" {
    type    = number
    default = 60000 // in ms
  }

  param "memory" {
    type    = number
    default = 128
  }

  param "memory_reservation" {
    type    = number
    default = 64
  }

  param "memory_swap" {
    type    = number
    default = 256
  }

  param "memory_swappiness" {
    type    = number
    default = 10
  }

  param "read_only" {
    type    = bool
    default = false
  }

  param "user" {
    type    = string
    default = "root"
  }

  step "container" "container_test_1" {
    image              = param.image
    cmd                = param.cmd
    env                = param.env
    timeout            = param.timeout
    memory             = param.memory
    memory_reservation = param.memory_reservation
    memory_swap        = param.memory_swap
    memory_swappiness  = param.memory_swappiness
    read_only          = param.read_only
    user               = param.user
  }

  output "stdout" {
    value = step.container.container_test_1.stdout
  }

  output "stderr" {
    value = step.container.container_test_1.stderr
  }

  output "exit_code" {
    value = step.container.container_test_1.exit_code
  }
}

pipeline "simple_container_step_with_param_override" {

  param "image" {
    type    = string
    default = "alpine:3.7"
  }

  param "cmd" {
    type    = list(string)
    default = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]
  }

  param "env" {
    type    = map(string)
    default = {
      FOO = "bar"
    }
  }

  param "timeout" {
    type    = number
    default = 60000 // in ms
  }

  param "memory" {
    type    = number
    default = 128
  }

  param "read_only" {
    type    = bool
    default = false
  }

  param "user" {
    type    = string
    default = "flowpipe"
  }

  step "container" "container_test_1" {
    image              = param.image
    cmd                = param.cmd
    env                = param.env
    timeout            = param.timeout
    memory             = param.memory
    read_only          = param.read_only
    user               = param.user
  }

  output "stdout" {
    value = step.container.container_test_1.stdout
  }

  output "stderr" {
    value = step.container.container_test_1.stderr
  }

  output "exit_code" {
    value = step.container.container_test_1.exit_code
  }
}

pipeline "simple_container_step_missing_image" {

  step "container" "container_test_1" {
    cmd = [
      "echo",
      "hello world"
    ]

    env = {
      FOO = "bar"
    }

    timeout = 60000 // in ms
    memory  = 128
  }
}

pipeline "simple_container_step_invalid_memory" {

  step "container" "container_test_1" {
    image = "alpine:3.7"

    cmd = [
      "echo",
      "hello world"
    ]

    env = {
      FOO = "bar"
    }

    timeout = 60000 // in ms
    memory  = 1
  }
}

pipeline "simple_container_step_with_string_timeout" {

  step "container" "container_test_1" {
    image = "alpine:3.7"

    cmd = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]

    env = {
      FOO = "bar"
    }

    timeout            = "60s"
    memory             = 128
    memory_reservation = 64
    memory_swap        = 256
    memory_swappiness  = 10
    read_only          = false
    user               = "root"
  }

  output "stdout" {
    value = step.container.container_test_1.stdout
  }

  output "stderr" {
    value = step.container.container_test_1.stderr
  }

  output "exit_code" {
    value = step.container.container_test_1.exit_code
  }

  output "lines" {
    value = step.container.container_test_1.lines[*]
  }
}

pipeline "simple_container_step_with_string_timeout_with_param" {

  param "timeout" {
    type    = string
    default = "60s"
  }

  step "container" "container_test_1" {
    image = "alpine:3.7"

    cmd = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]

    env = {
      FOO = "bar"
    }

    timeout            = param.timeout
    memory             = 128
    memory_reservation = 64
    memory_swap        = 256
    memory_swappiness  = 10
    read_only          = false
    user               = "root"
  }

  output "stdout" {
    value = step.container.container_test_1.stdout
  }

  output "stderr" {
    value = step.container.container_test_1.stderr
  }

  output "exit_code" {
    value = step.container.container_test_1.exit_code
  }

  output "lines" {
    value = step.container.container_test_1.lines[*]
  }
}