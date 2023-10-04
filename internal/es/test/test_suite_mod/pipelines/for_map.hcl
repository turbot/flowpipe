pipeline "for_map" {
    param "legends" {
        type = map

        default = {
            "janis" = {
                last_name= "joplin"
                age = 27
            }
            "jimi" = {
                last_name= "hendrix"
                age = 27
            }
            "jerry" = {
                last_name= "garcia"
                age = 53
            }
        }
    }

    step "echo" "text_1" {
        for_each = param.legends
        text = "${each.key} ${each.value.last_name} was ${each.value.age}"
    }

    output "text_1" {
        value = step.echo.text_1[0].text
    }

    output "text_2" {
        value = step.echo.text_1[1].text
    }

    output "text_3" {
        value = step.echo.text_1[2].text
    }
}