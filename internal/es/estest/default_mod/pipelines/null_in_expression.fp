pipeline "hn_parent" {

    param "stories" {
        type = any
        default = [
            {
                "id": 39689817,
                "text": "Has anyone else found setting up an Azure account to be particularly challenging compared to AWS? Our multinational company initially created our Azure account outside the EU, and we&#x27;ve faced numerous hurdles trying to move it to an EU state. Issues ranged from unrecognized credit cards to confusion over subscriptions, admin roles, and permission settings. It feels unnecessarily complex. AWS seems to encourage creating multiple interconnected accounts, a stark contrast to Azure's limitations. Is this a common experience, or are we missing something? Appreciate any insights or tips.",
                "time": "2024-03-13T10:21:51.000Z",
                "title": "Ask HN: Why Is Setting Up an Azure Account More Difficult Than AWS?",
                "url": null
            }
        ]
    }

    step "pipeline" "call_hn" {
        pipeline = pipeline.hn_top
        args = {
            stories = param.stories
        }
    }
}

pipeline "hn_top" {
    param "stories" {
        type = any
    }

  step "transform" "post_message_to_slack" {
    for_each = param.stories
    value = try(each.value.url, "https://news.ycombinator.com/item?id=${each.value.id}")
  }
}