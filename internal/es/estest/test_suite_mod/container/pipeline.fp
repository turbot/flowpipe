pipeline "container_from_source" {

  step "container" "container_test_1" {
    source = "./container/Dockerfile"

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
    value = step.container.container_test_1.lines
  }
}
