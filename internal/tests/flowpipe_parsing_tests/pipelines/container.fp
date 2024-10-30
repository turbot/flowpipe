pipeline "pipeline_step_container" {

  description = "Container step test pipeline"

  step "container" "container_test1" {
    image              = "test/image"
    cmd                = ["foo", "bar"]
    entrypoint         = ["foo", "baz"]
    timeout            = 60000 // in ms
    memory             = 128
    memory_reservation = 64
    memory_swap        = -1
    memory_swappiness  = 60
    cpu_shares         = 512
    read_only          = false
    user               = "flowpipe"
    workdir            = "."
    env = {
      ENV_TEST = "hello world"
    }
  }
}

pipeline "pipeline_step_with_param" {

  description = "Container step test pipeline"

  param "region" {
    description = "The name of the region."
    type        = string
    default     = "ap-south-1"
  }

  param "image" {
    description = "The name of the image."
    type        = string
    default     = "test/image"
  }

  param "cmd" {
    description = "The list of the commands to be run."
    type        = list(string)
    default     = ["foo", "bar"]
  }

  param "entry_point" {
    description = "Entrypoint of the image."
    type        = list(string)
    default     = ["foo", "bar", "baz"]
  }

  param "timeout" {
    description = "The timeout of the container run."
    type        = number
    default     = 120000 // in ms
  }

  param "cpu_shares" {
    description = "CPU shares (relative weight) for the container."
    type        = number
    default     = 512
  }

  param "memory" {
    description = "Amount of memory in MB your container can use at runtime."
    type        = number
    default     = 128
  }

  param "memory_reservation" {
    description = "Specify a soft limit smaller than the memory."
    type        = number
    default     = 64
  }

  param "memory_swap" {
    description = "The amount of memory this container is allowed to swap to disk."
    type        = number
    default     = -1
  }

  param "memory_swappiness" {
    description = "Tune container memory swappiness (0 to 100)."
    type        = number
    default     = 60
  }

  param "read_only" {
    description = "If true, the container will be started as readonly."
    type        = bool
    default     = true
  }

  param "container_user" {
    description = "User to run the container."
    type        = string
    default     = "flowpipe"
  }

  param "work_dir" {
    description = "The working directory for commands to run in."
    type        = string
    default     = "."
  }

  step "container" "container_test1" {
    image              = param.image
    cmd                = param.cmd
    entrypoint         = param.entry_point
    timeout            = param.timeout
    cpu_shares         = param.cpu_shares
    memory             = param.memory
    memory_reservation = param.memory_reservation
    memory_swap        = param.memory_swap
    memory_swappiness  = param.memory_swappiness
    read_only          = param.read_only
    user               = param.container_user
    workdir            = param.work_dir
    env = {
      REGION = param.region
    }
  }
}

pipeline "pipeline_step_container_timeout_string" {

  description = "Container step test pipeline"

  step "container" "container_test1" {
    image              = "test/image"
    cmd                = ["foo", "bar"]
    entrypoint         = ["foo", "baz"]
    timeout            = "60s"
    memory             = 128
    memory_reservation = 64
    memory_swap        = -1
    memory_swappiness  = 60
    cpu_shares         = 512
    read_only          = false
    user               = "flowpipe"
    workdir            = "."
    env = {
      ENV_TEST = "hello world"
    }
  }
}

pipeline "pipeline_step_container_with_param_timeout_string" {

  description = "Container step test pipeline"

  param "region" {
    description = "The name of the region."
    type        = string
    default     = "ap-south-1"
  }

  param "image" {
    description = "The name of the image."
    type        = string
    default     = "test/image"
  }

  param "cmd" {
    description = "The list of the commands to be run."
    type        = list(string)
    default     = ["foo", "bar"]
  }

  param "entry_point" {
    description = "Entrypoint of the image."
    type        = list(string)
    default     = ["foo", "bar", "baz"]
  }

  param "timeout" {
    description = "The timeout of the container run."
    type        = string
    default     = "120s"
  }

  param "cpu_shares" {
    description = "CPU shares (relative weight) for the container."
    type        = number
    default     = 512
  }

  param "memory" {
    description = "Amount of memory in MB your container can use at runtime."
    type        = number
    default     = 128
  }

  param "memory_reservation" {
    description = "Specify a soft limit smaller than the memory."
    type        = number
    default     = 64
  }

  param "memory_swap" {
    description = "The amount of memory this container is allowed to swap to disk."
    type        = number
    default     = -1
  }

  param "memory_swappiness" {
    description = "Tune container memory swappiness (0 to 100)."
    type        = number
    default     = 60
  }

  param "read_only" {
    description = "If true, the container will be started as readonly."
    type        = bool
    default     = true
  }

  param "container_user" {
    description = "User to run the container."
    type        = string
    default     = "flowpipe"
  }

  param "work_dir" {
    description = "The working directory for commands to run in."
    type        = string
    default     = "."
  }

  step "container" "container_test1" {
    image              = param.image
    cmd                = param.cmd
    entrypoint         = param.entry_point
    timeout            = param.timeout
    cpu_shares         = param.cpu_shares
    memory             = param.memory
    memory_reservation = param.memory_reservation
    memory_swap        = param.memory_swap
    memory_swappiness  = param.memory_swappiness
    read_only          = param.read_only
    user               = param.container_user
    workdir            = param.work_dir
    env = {
      REGION = param.region
    }
  }
}
