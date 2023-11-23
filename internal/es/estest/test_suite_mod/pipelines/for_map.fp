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

    step "transform" "text_1" {
        for_each = param.legends
        value    = "${each.key} ${each.value.last_name} was ${each.value.age}"
    }

    output "text_1" {
        value = step.transform.text_1["janis"].value
    }

    output "text_2" {
        value = step.transform.text_1["jimi"].value
    }

    output "text_3" {
        value = step.transform.text_1["jerry"].value
    }
}