pipeline "expr_depend_and_function" {
    step "echo" "text_1" {
        text = "foo"
    }

    step "echo" "text_2" {
        text = "lower case here ${title("bar ${step.echo.text_1.text} baz")} lower case here again"
    }

    step "echo" "text_3" {
        text = "output2 ${title(step.echo.text_2.text)} title(output1) ${title(step.echo.text_1.text)}"
    }
}


pipeline "expr_list_to_json" {
    step "echo" "text_1" {
        json = tolist(["foo", "bar", "baz"])
    }
}

# pipeline "expr_list_to_dynamic" {
#     step "echo" "text_1" {
#         dynamic = tolist(["foo", "bar", "baz"])
#     }
# }