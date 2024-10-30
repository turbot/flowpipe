mod "complex_variable" {
 
}


variable "base_tag_rules" {
  type = object({
    add           = optional(map(string))
    remove        = optional(list(string))
    remove_except = optional(list(string))
    update_keys   = optional(map(list(string)))
    update_values = optional(map(map(list(string))))
  })
  description = "Base rules to apply to resources unless overridden when merged with any provided resource-specific rules."
  default     = {
    add           = {}
    remove        = []
    remove_except = []
    update_keys   = {}
    update_values = {}
  }
}

variable "accessanalyzer_tag_rules" {
  type = object({
    add           = optional(map(string))
    remove        = optional(list(string))
    remove_except = optional(list(string))
    update_keys   = optional(map(list(string)))
    update_values = optional(map(map(list(string))))
  })
  description = "Access Analyzers specific tag rules"
  default     = null
}

pipeline "foo" {
    step "transform" "echo" {
        value = var.base_tag_rules
    }
}
