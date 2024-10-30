mod "test_mod" {
  title = "my_mod"
}

locals {
  artists = [
    {
      name  = "Real Friends",
      album = "Maybe This Place Is The Same And We're Just Changing",
    },
    {
      name  = "A Day To Remember",
      album = "Common Courtesy",
    },
    {
      name  = "The Story So Far",
      album = "What You Don't See",
    }
  ]
}

pipeline "with_step_output" {

  step "transform" "name" {
    for_each = local.artists
    value    = "artist name: ${each.value.name}"

    output "album_name" {
      value = "album name: ${each.value.album}"
    }
  }

  step "transform" "second_step" {
    for_each = step.transform.name
    value    = "album name: ${each.value.output.album_name}"
  }
}
