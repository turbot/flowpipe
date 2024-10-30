mod "test" {

}

pipeline "max_function" {
    param "stories" {
        type = any
    }

    step "transform" "post_message_to_slack" {
        for_each = param.stories

        value = max(5, each.value.url)
    }
}

pipeline "try_function" {
    param "stories" {
        type = any
    }

    step "transform" "try_func" {
        for_each = param.stories

        // try function generates a different error message, need to ensure that it can be handled
        value = try(each.value.url, "https://news.ycombinator.com/item?id=${each.value.id}")
    }
}

pipeline "try_function_no_for_each" {
    param "my_param" {
        type = any
    }

    step "transform" "first" {
        value = param.my_param
    }

    step "transform" "second" {
        // try function generates a different error message, need to ensure that it can be handled
        value = try(step.transform.first.value, "foo")
    }
}

pipeline "try_function_no_for_each_combination_1" {
    param "my_param" {
        type = any
    }

    step "transform" "first" {
        value = param.my_param
    }

    step "transform" "second" {
        // try function generates a different error message, need to ensure that it can be handled
        value = max(try(step.transform.first.value, 4), 5)
    }
}

pipeline "try_function_no_for_each_combination_2" {
    param "my_param" {
        type = any
    }

    step "transform" "first" {
        value = param.my_param
    }

    step "transform" "number" {
        value = 5
    }

    step "transform" "second" {
        // try function generates a different error message, need to ensure that it can be handled
        value = max(try(step.transform.first.value, 4), 5, step.transform.number.value)
    }
}

pipeline "try_function_within_json_encode" {
  param "slack_webhook_url" {
    type        = string
    description = "The webhook URL for the Slack channel."
   #  default     = "https://hooks.slack.com/services/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"
  }

  param "stories" {
    type = any
  }

  step "transform" "nexus" {
    value = param.stories
  }

  step "http" "post_message_to_slack" {
    for_each = param.stories

    method = "post"
    url    = param.slack_webhook_url

    request_headers = {
      Content-Type = "application/json"
    }

    request_body = jsonencode({
      attachments = [
        {
          color = "#ff6600"
          fallback = "${each.value.title}"
          fallback = step.transform.nexus.value
          title = each.value.title
          text = each.value.text
          title_link = try(each.value.url, "https://news.ycombinator.com/item?id=${each.value.id}")
          author_name = "#${each.value.id}"
          author_link = "https://news.ycombinator.com/item?id=${each.value.id}"
          footer = "Hacker News"
        }
      ]
    })
  }
}


pipeline "try_function_from_param" {
    param "my_param" {
        type = any
    }

    step "transform" "try_func" {
        // having a param automatically add as "unresolved"
        value = try(param.my_param.value, "foo")
    }
}
