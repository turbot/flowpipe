
mod "trigger_mod" {
  title = "Mod with Triggers"
    require {
        mod "mod_depend_a" {
            version = "1.0.0"
        }
        mod "mod_depend_b" {
            version = "1.0.0"
        }
    }
}