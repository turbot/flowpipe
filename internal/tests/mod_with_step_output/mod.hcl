mod "test_mod" {
  title = "my_mod"
}

locals {
    artists = [
        {
            name = "Real Friends",
            album = "Maybe This Place Is The Same And We're Just Changing",
        },
        {
            name = "A Day To Remember",
            album = "Common Courtesy",
        },
        {
            name = "The Story So Far",
            album = "What You Don't See",
        }
    ]
}

pipeline "with_step_output" {

    step "echo" "name" {
        for_each = local.artists
        text = "artist name: ${each.value.name}"

        output "album_name" {
            value = "album name: ${each.value.album}"
        }
    }

    step "echo" "second_step" {
        for_each = step.echo.name
        text = "album name: ${each.value.output.album_name}"
    }
}