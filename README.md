# Flowpipe

Flowpipe is "pipelines as code", defining workflows and other tasks
that are performed in a sequence.

Flowpipe consists of:
* Triggers - A way to initiate a pipeline (e.g. cron, webhook, etc)
* Pipelines - A sequence of steps to run actions
* Service - Manage triggers and execute pipelines

## Examples

### Call a webhook every hour [trigger.interval, step.http_request]

```hcl
trigger "interval" "hourly" {
    interval = "hourly"
    pipeline = pipeline.job
}

pipeline "job" {
    step "http_request" "req" {
        url = "https://example.com/my/webhook"
    }
}
```

### Run a python function on calls to a webhook [trigger.http, step.function]

```hcl
trigger "http" "my_webhook" {
    pipeline = pipeline.my_webhook_pipeline
}

pipeline "my_webhook_pipeline" {
    step "function" "my_python_function" {
        location = "functions/add_owner_tag_to_instance"
    }
}
```

### Run a python function for events [param]

```hcl
trigger "http" "my_webhook" {
    pipeline = pipeline.my_webhook_pipeline
    args = {
        event = self.payload
    }
}

pipeline "my_webhook_pipeline" {

    param "event" {
        type = "object"
    }

    step "function" "my_python_function" {
        location = "functions/add_owner_tag_to_instance"
        # TODO - Match lambda signature? e.g. event or input?
        input = param.event
    }
}
```

### Run a python function for events of a given type [if]

```hcl
trigger "http" "my_webhook" {
    pipeline = pipeline.my_webhook_pipeline
    args = {
        event = self.payload
    }
}

pipeline "my_webhook_pipeline" {

    param "event" {
        type = "object"
    }

    step "function" "my_python_function" {
        if = param.event.type == "aws_instance"
        location = "functions/aws_instance_function"
    }

    step "function" "my_other_python_function" {
        if = param.event.type != "aws_instance"
        location = "functions/other_function"
    }

    # TODO - can we refer to the result of the other if to avoid duplicate logic?
    step "function" "my_other_python_function" {
        if = !step.function.my_python_function
        location = "functions/other_function"
    }

}
```

### Choose the python function for events based on their type [dynamic function location]

```hcl
trigger "http" "my_webhook" {
    pipeline = pipeline.my_webhook_pipeline
    args = {
        event = self.response_body
    }
}

pipeline "my_webhook_pipeline" {

    param "event" {
        type = "object"
    }

    step "function" "my_python_function" {
        # TODO - is location a runtime or compile time attribute?
        location = "functions/${param.event.type}"
    }
}
```

### Function switcher [directives]

```hcl
trigger "http" "my_webhook" {
    pipeline = pipeline.my_webhook_pipeline
    args = {
        event = self.response_body
    }
}

pipeline "my_webhook_pipeline" {

    param "event" {
        type = "object"
    }

    step "function" "my_python_function" {
        location = "functions/%{if param.event.type == "aws_instance" }aws_instance%{else}other%{endif}"
    }
}
```

### For loop to send new AWS IAM access keys to Slack [trigger.query, pipeline params, for_each]

```hcl
trigger "query" "recent_mentions_on_twitter" {
    sql = <<EOQ
        select
            access_key_id,
            user_name,
            account_id
        from
            aws_iam_access_key;
    EOQ

    # Only run the pipeline when keys are discovered to be expired
    events = [ "insert" ]

    # Not needed, would be unique for each row anyway
    primary_key = "access_key_id"

    pipeline = pipeline.send_attachments_to_slack
    args = {
        access_keys = self.rows
    }
}

pipeline "send_keys_to_slack" {

    param "access_keys" {
        type = "list(object)"
    }

    step "http_request" "send_to_slack" {
        for_each = param.access_keys
        url = var.slack_webhook_url
        body = jsonencode({
            text = "New AWS IAM access key: ${each.value.access_key_id} for ${each.value.user_name} in ${each.value.account_id}"
        })
    }

}
```


### For loop to send new Active AWS IAM access keys to Slack [for_each is before if]

* for_each is evaluated before if
* to do an if on the entire block use for_each = condition ? items : []

```hcl
trigger "query" "recent_mentions_on_twitter" {
    sql = <<EOQ
        select
            access_key_id,
            user_name,
            status,
            account_id
        from
            aws_iam_access_key;
    EOQ

    # Only run the pipeline when keys are discovered to be expired
    events = [ "insert" ]

    # Not needed, would be unique for each row anyway
    primary_key = "access_key_id"

    pipeline = pipeline.send_attachments_to_slack
    args = {
        access_keys = self.rows
    }
}

pipeline "send_keys_to_slack" {

    param "access_keys" {
        type = "list(object)"
    }

    step "http_request" "send_to_slack" {
        for_each = param.access_keys
        if = each.value.status == "Active"
        url = var.slack_webhook_url
        body = jsonencode({
            text = "New AWS IAM access key: ${each.value.access_key_id} for ${each.value.user_name} in ${each.value.account_id}"
        })
    }

}
```

### Call a pipeline from within a pipeline [step.pipeline, depends_on]

```hcl
trigger "http" "my_webhook" {
    pipeline = pipeline.my_webhook_pipeline
}

pipeline "my_webhook_pipeline" {
    step "pipeline" "run_1" {
        pipeline = pipeline.my_reusable_pipeline
    }
    step "function" "run_2" {
        depends_on = [ step.pipeline.run_1 ]
        pipeline = pipeline.my_reusable_pipeline
    }
}

pipeline "my_reusable_pipeline" {
    step "function" "my_python_function" {
        location = "functions/add_owner_tag_to_instance"
    }
}
```

### Pipeline output and params [output]

```hcl
trigger "http" "my_webhook" {
    pipeline = pipeline.my_webhook_pipeline
    args = {
        event = self.response_body
    }
}

pipeline "my_webhook_pipeline" {
    param "event" {
        type = "object"
    }
    step "pipeline" "part_1" {
        pipeline = pipeline.pipeline_1
        args = {
            event = param.event
        }
    }
    step "http" "part_2" {
        body = step.pipeline.part_1.output.data
    }
}

pipeline "pipeline_1" {
    param "event" {
        type = "map"
    }
    step "function" "my_python_function" {
        location = "functions/add_owner_tag_to_instance"
        input = param.event
    }
    output "data" {
        # TODO should the result of a function be in an attribute called output or value or ???
        value = step.function.my_python_function.output
    }
}
```

### Delete AWS access keys older than 90 days [step.container, connection context]

```hcl
trigger "query" "expired_access_keys" {
    sql = <<EOQ
        select
            access_key_id,
            user_name,
            create_date,
            ctx ->> 'connection_name' as connection
        from
            aws_iam_access_key
        where
            create_date < now() - interval '90 days'
    EOQ

    # Only run the pipeline when keys are discovered to be expired
    events = [ "insert" ]
    primary_key = "access_key_id"

    pipeline = pipeline.delete_expired_access_keys
    args = {
        access_keys = self.rows
    }
}

pipeline "delete_expired_access_keys" {

    param "access_keys" {
        type = "list(object)"
    }

    step "container" "delete_access_key" {
        # The step should be run for every row in the input
        for_each = param.access_keys

        # Use the connection for the row
        connection = each.value.connection

        # Call the AWS CLI
        image = "amazon/aws-cli"
        cmd = ["aws", "iam", "delete-access-key", "--user-name", each.value.user_name, "--access-key-id", each.value.access_key_id]
        # Or shell form (be careful with quoting)
        cmd_shell = "aws iam delete-access-key --user-name ${each.value.user_name} --access-key-id ${each.value.access_key_id}"
    }
}
```

### Send new Tweets to Slack

```hcl
trigger "query" "recent_mentions_on_twitter" {
    sql = <<EOQ
        select
            '#1DA1F2' as color,
            'https://twitter.com/' || (author->>'username') || '/statuses/' || id as fallback,
            text,
            id || ' by @' || (author->>'username') as title,
            'https://twitter.com/' || (author->>'username') || '/statuses/' || id as title_link,
            'Twitter' as footer
        from
            twitter_search_recent
        where
            query = '(steampipe OR steampipe.io OR github.com/turbot) -tomathy -GutturalSteve -"steampipe alley"'
    EOQ

    # Only run the pipeline when keys are discovered to be expired
    events = [ "insert" ]
    primary_key = "fallback"

    pipeline = pipeline.send_attachments_to_slack
    args = {
        attachments = self.rows
    }
}

pipeline "send_attachments_to_slack" {

    param "attachments" {
        type = "list(object)"
    }

    step "http_request" "send_to_slack" {
        url = var.slack_webhook_url
        body = jsonencode({
            attachments = param.attachments
        })
    }

}
```

### Send new Tweets to Slack with transform step [step.transform]

```hcl
trigger "query" "recent_mentions_on_twitter" {
    sql = <<EOQ
        select 
            id,
            text,
            'https://twitter.com/' || (author->>'username') || '/statuses/' || id as url,
            author->>'username' as username,
            created_at
        from
            twitter_search_recent
        where
            query = '(steampipe OR steampipe.io OR github.com/turbot) -tomathy -GutturalSteve -"steampipe alley"'
    EOQ

    # Only run the pipeline when keys are discovered to be expired
    events = [ "insert" ]
    primary_key = "fallback"

    pipeline = pipeline.send_tweets_to_slack
    args = {
        tweets = self.output.rows
    }
}

pipeline "send_tweets_to_slack" {

    param "tweets" {
        type = "list"
    }

    step "transform" "slack_attachments" {
        output = [
            for tweet in param.tweets : {
                color = "#1DA1F2"
                fallback = tweet.url
                text = tweet.text
                title = "${tweet.id} by @${tweet.username}"
                title_link = tweet.url
                footer = "Twitter"
            }
        ]
    }

    step "http_request" "send_to_slack" {
        url = var.slack_webhook_url
        body = jsonencode({
            attachments = step.transform.slack_attachments.output
        })
    }

}
```

### Send new Tweets to Slack one at a time

```hcl
trigger "query" "recent_mentions_on_twitter" {
    sql = <<EOQ
        select 
            id,
            text,
            'https://twitter.com/' || (author->>'username') || '/statuses/' || id as url,
            author->>'username' as username,
            created_at
        from
            twitter_search_recent
        where
            query = '(steampipe OR steampipe.io OR github.com/turbot) -tomathy -GutturalSteve -"steampipe alley"'
    EOQ

    # Only run the pipeline when keys are discovered to be expired
    events = [ "insert" ]
    primary_key = "fallback"

    pipeline = pipeline.send_tweets_to_slack
    args = {
        tweets = self.output.rows
    }
}

pipeline "send_tweets_to_slack" {

    param "slack_webhook_url" {
        type    = "string"
        default = var.slack_webhook_url
    }

    param "tweets" {
        type = "list"
    }

    step "transform" "slack_attachments" {
        output = [
            for tweet in param.tweets : {
                color = "#1DA1F2"
                fallback = tweet.url
                text = tweet.text
                title = "${tweet.id} by @${tweet.username}"
                title_link = tweet.url
                footer = "Twitter"
            }
        ]
    }

    step "http_request" "send_to_slack" {
        for_each = step.transform.slack_attachments.output
        url = param.slack_webhook_url
        body = jsonencode({
            attachments = [each.value]
        })
    }

}
```

### Depending on results of a for_each step [execution flow with for_each]

```hcl
trigger "interval" "hourly" {
    interval = "hourly"
    pipeline = pipeline.job
}

pipeline "job" {

    step "query" "keys" {
        sql = <<EOQ
            select
                access_key_id,
                user_name,
                account_id
            from
                aws_iam_access_key;
        EOQ
    }

    step "http_request" "send_to_slack" {
        for_each = param.access_keys
        url = var.slack_webhook_url
        payload = jsonencode({
            text = "AWS IAM access key: ${each.value.access_key_id} for ${each.value.user_name} in ${each.value.account_id}"
        })
    }

    # This step will only run when the previous step is complete
    step "http_request" "send_summary_to_slack" {
        url = var.slack_webhook_url
        payload = jsonencode({
            text = "Total keys sent: ${length(step.http_request.send_to_slack)}"
        })
    }

    # Or don't wait for separate messages
    step "http_request" "send_summary_to_slack" {
        url = var.slack_webhook_url
        payload = jsonencode({
            text = "Total keys sent: ${length(step.query.keys.rows)}"
        })
    }

    # Lookup the result of a specific step in the loop
    step "http_request" "send_first_result_to_slack" {
        url = var.slack_webhook_url
        payload = jsonencode({
            text = "First key sent: ${step.http_request.send_to_slack[0].body}"
        })
    }

}
```

### Execute multiple steps for each item in a for_each

```hcl
trigger "interval" "hourly" {
    interval = "hourly"
    pipeline = pipeline.job
}

pipeline "job" {

    step "query" "keys" {
        sql = <<EOQ
            select
                access_key_id,
                user_name,
                account_id,
                _ctx ->> 'connection_name' as connection
            from
                aws_iam_access_key;
        EOQ
    }

    # These run in parallel by default
    step "pipeline" "handle_key" {
        for_each = param.access_keys
        pipeline = pipeline.handle_key
        args = {
            access_key = each.value
        }
    }

    # This step will only run when the previous step is complete
    step "http_request" "send_summary_to_slack" {
        url = var.slack_webhook_url
        payload = jsonencode({
            text = "Keys handled: ${length(step.pipeline.handle_key)}"
        })
    }

    # Alternative - depends_on
    step "http_request" "send_summary_to_slack" {
        depends_on = [step.pipeline.handle_key]
        url = var.slack_webhook_url
        payload = jsonencode({
            text = "Keys handled"
        })
    }

}

pipeline "handle_key" {

    param "access_key" {
        type = "object"
    }

    step "http_request" "send_warning_to_slack" {
        url = var.slack_webhook_url
        payload = jsonencode({
            text = "Deleting key: ${param.access_key_id} for ${param.user_name} in ${param.account_id}"
        })
    }

    step "container" "delete_access_key" {
        # Use the connection for the row
        connection = param.access_key.connection
        image = "amazon/aws-cli"
        cmd = ["aws", "iam", "delete-access-key", "--user-name", param.user_name, "--access-key-id", param.access_key_id]
    }

    step "http_request" "send_to_slack" {
        url = var.slack_webhook_url
        payload = jsonencode({
            title = "Deleted key: ${param.access_key_id} for ${param.user_name} in ${param.account_id}"
            text = step.container.delete_access_key.stdout
        })
    }

}
```

### Send new Tweets using a cron and deduplication

```hcl
trigger "interval" "hourly" {
    interval = "hourly"

    pipeline = pipeline.twitter_mentions_to_slack
    args = {
        query = "(steampipe OR steampipe.io OR github.com/turbot) -tomathy -GutturalSteve -\"steampipe alley\""
    }

}

pipeline "send_tweets_to_slack" {

    param "slack_webhook_url" {
        type    = "string"
        default = var.slack_webhook_url
    }

    param "query" {
        type = "string"
    }

    step "query" "recent_mentions_on_twitter" {
        sql = <<EOQ
            select 
                id,
                text,
                'https://twitter.com/' || (author->>'username') || '/statuses/' || id as url,
                author->>'username' as username,
                created_at
            from
                twitter_search_recent
            where
                query = $1
        EOQ
        args = {
            query = param.query
        }
    }

    step "dedup" "new_recent_mentions_on_twitter" {
        input = step.query.recent_mentions_on_twitter.output.rows
        primary_key = [ "id" ]
    }

    step "transform" "slack_attachments" {
        output = [
            for tweet in step.dedup.new_recent_mentions_on_twitter : {
                color = "#1DA1F2"
                fallback = tweet.url
                text = tweet.text
                title = "${tweet.id} by @${tweet.username}"
                title_link = tweet.url
                footer = "Twitter"
            }
        ]
    }

    step "http_request" "send_to_slack" {
        url = param.slack_webhook_url
        body = jsonencode({
            attachments = step.transform.slack_attachments.output
        })
    }

}
```



## Triggers

Triggers are the way to initiate a pipeline. They are defined in the
mod and are based on a schedule, webhook or other event.

For example:
```hcl
# A webhook trigger will automatically generate a webhook URL. The URL accepts
# a JSON payload and passes it as an input to the defined pipeline.
trigger "http" "my_webhook" {
    # Pipeline to execute [required].
    pipeline = pipeline.my_webhook_pipeline
    # Defaults, not needed for basic webhook
    # method = "post"
    # content_type = "application/json"
}

# A form trigger will automatically generate a form URL. The URL accepts
# a form encoded payload and passes it as an input to the defined pipeline.
trigger "http" "my_form" {
    # Pipeline to execute [required].
    pipeline = pipeline.my_webhook_pipeline
    # Set the content type for a form
    content_type = "multipart/form-data"
}

# The query trigger will execute a SQL query on a schedule and pass row changes
# as an input to the defined pipeline.
trigger "query" "my_query" {
    # Pipeline to execute [required].
    pipeline = pipeline.my_account_handler
    # SQL query to execute [required].
    sql = "select id, title from aws_account"
    # Schedule to run the query [optional, default hourly].
    schedule = "* * * * * *"
    # Events to trigger on [optional, default insert, update, delete].
    events = [ "insert", "update", "delete" ]
    # Primary key to use for update vs insert detection [optional].
    primary_key = "id"
    # How to map query results to pipeline inputs [optional].
    mapping = "pipeline_per_row" # or "pipeline_per_query"
}

# The cron trigger fires off an event on the given schedule. This time event is
# passed as input to the defined pipeline.
trigger "cron" "my_cron" {
    # Pipeline to execute [required].
    pipeline = pipeline.my_scheduled_pipeline
    # Schedule to run the query [optional, default hourly].
    schedule = "* * * * * *"
}
```

Alternate syntax?

```hcl
# Top level like dashboard components
http "my_webhook" {
    pipeline = pipeline.my_webhook_pipeline
}

# Similar to connection definitions
trigger "my_webhook" {
    type = "http"
    pipeline = pipeline.my_webhook_pipeline
}
```

## Pipelines

### Using primitives

Single step pipeline running a HTTP request passing a JSON payload:
```hcl
pipeline "my_webhook_pipeline" {
    step "my_http_request" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # method = "post"
        # content_type = "application/json"
        # TODO - does this need to be a JSON string for diff types inside the map?
        body = {
            my_input = "my_value"
        }
    }
}
```

Single step pipeline running a HTTP form passing a JSON payload:
```hcl
pipeline "my_form_pipeline" {
    step "my_form_submit" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # method = "post"
        content_type = "multipart/form-data"
        # convert to form fields based on the content type?
        body = {
            my_input = "my_value"
        }
    }
}
```

Single step pipeline to run an AWS CLI command. Flowpipe will inspect the location
and automatically create the correct Lambda function container to compile and wrap
it, in this case, a shell script:
```hcl
pipeline "my_aws_cli" {
    step "add_owner_tag_to_instance" {
        base = steampipe.step.function
        # location = "functions/add_owner_tag_to_instance"
        input = {
            Owner = "Nathan"
        }
    }
}
```

As above, but run a boto3 python script. No change to the HCL, just the path
has python code rather than a shell script:
```hcl
pipeline "my_boto3_pipeline" {
    step "add_owner_tag_to_instance" {
        base = steampipe.step.function
        # location = "functions/add_owner_tag_to_instance"
        input = {
            Owner = "Nathan"
        }
    }
}
```

Run a Steampipe SQL query:
```hcl
pipeline "my_steampipe_pipeline" {
    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
        # Optional, defaults
        # connection_string = "postgres://steampipe@localhost:9193/steampipe"
    }
}
```

Run a Steampipe query, using a base query and passing args to it:
```hcl
pipeline "my_steampipe_pipeline" {
    step "my_query" {
        base = steampipe.step.query
        # connection_string = "postgres://steampipe@localhost:9193/steampipe"
        query = query.my_tag_filter_query
        args = {
            tag_name = "Owner"
        }
    }
}
```

### Chaining steps

Prompt for input, with a chained step:
```hcl
pipeline "my_chained_pipeline" {

    step "my_prompt" {
        # TODO - should this use the input like a dashboard?
        base = steampipe.step.prompt
        title = "Owner name"
        type = "string"
        placeholder = "Jane Doe"
        # default = "Jane Doe"
    }

    step "my_http_request_with_input_chain" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        body = {
            owner = step.my_prompt.output.value
        }
    }

}
```

### Terraform style control flow

#### Loops

Looping through results of a query:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
    }

    step "my_webhook_loop" {
        for_each = step.my_query.output.rows
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        # input = each.value
        # Or, get a single value out to pass
        input = {
            "account_id" = each.value.account_id
        }
    }

}
```

Independent, parallel loops through results of a query:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
    }

    step "my_webhook_loop" {
        for_each = step.my_query.output.rows
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        # input = each.value
        # Or, get a single value out to pass
        input = {
            "account_id" = each.value.account_id
        }
    }

    step "my_parallel_webhook_loop" {
        for_each = step.my_query.output.rows
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint2"
        # Pass JSON for each row directly
        # input = each.value
        # Or, get a single value out to pass
        input = {
            "account_id" = each.value.account_id
        }
    }

}
```

Loop through results of a query, and then do a follow on step for each item:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
    }

    step "my_webhook_loop" {
        for_each = step.my_query.output.rows
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        # input = each.value
        # Or, get a single value out to pass
        input = {
            "account_id" = each.value.account_id
        }
    }

    step "my_parallel_webhook_loop" {
        # Create a step flow for each step flow above
        for_each = step.my_webhook_loop
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint2"
        # Use the request body from the first my_webhook_loop step
        input = each.value.body
    }

}
```

#### IF/ELSE/CASE via Terraform style count

IF via count:
```hcl
pipeline "my_steampipe_pipeline" {

    # Only called if the condition matched
    step "do_for_insert" {
        # Refers to the pipeline input
        count = self.input.type == "insert" ? 1 : 0
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.my_filter.output
    }

}
```

IF/ELSE via count:
```hcl
pipeline "my_steampipe_pipeline" {

    step "do_for_insert" {
        # Refers to the pipeline input
        count = self.input.type == "insert" ? 1 : 0
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = self.input
    }

    step "do_for_not_insert" {
        # Refers to the pipeline input
        count = self.input.type == "insert" ? 0 : 1
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = self.input
    }

}
```

CASE via count:
```hcl
pipeline "my_steampipe_pipeline" {

    step "do_for_insert" {
        # Refers to the pipeline input
        count = self.input.type == "insert" ? 1 : 0
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = self.input
    }

    step "do_for_update" {
        # Refers to the pipeline input
        count = self.input.type == "update" ? 0 : 1
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = self.input
    }

}
```

#### IF/ELSE/CASE with "if =" instead of "count ="

Option 1 - IF via count:
```hcl
pipeline "my_steampipe_pipeline" {

    # Option 1 - if =
    step "do_for_insert" {
        # Refers to the pipeline input
        if = self.input.type == "insert"
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.my_filter.output
    }

    # Option 1 - when =
    step "do_for_insert" {
        # Refers to the pipeline input
        when = self.input.type == "insert"
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.my_filter.output
    }

    # Option 3 - case {}
    step "do_for_insert" {
        # Refers to the pipeline input
        case {
            test = self.input.type
            when "insert" {
                # ugh ... not sure what to do here, perhaps embedded steps?
            }
        }
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.my_filter.output
    }

}
```

#### IF/ELSE/CASE blocks

IF step:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_if" {
        base = steampipe.step.if
        if = self.input.type == "insert"
        then {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = step.my_filter.output
        }
    }

}
```

IF/ELSE step:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_if" {
        base = steampipe.step.if
        if = self.input.type == "insert"
        then {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint_if"
            # Pass JSON for each row directly
            input = self.input
        }
        else {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint_else"
            # Pass JSON for each row directly
            input = self.input
        }
    }

}
```

CASE step - string:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_if" {
        base = steampipe.step.case
        test = self.input.type
        when "insert" {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint_insert"
            # Pass JSON for each row directly
            input = self.input
        }
        when "update" {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint_update"
            # Pass JSON for each row directly
            input = self.input
        }
        else {
            base = steampipe.step.http_request
            # dynamic works!
            url = "https://example.com/with/endpoint_${self.input.type}"
            # Pass JSON for each row directly
            input = self.input
        }
    }

}
```

CASE step - non-string:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_if" {
        base = steampipe.step.case
        when {
            # The test can be put inside the when block, they are run in order
            test = self.input.count > 24
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint_large"
            # Pass JSON for each row directly
            input = self.input
        }
        else {
            base = steampipe.step.http_request
            # dynamic works!
            url = "https://example.com/with/endpoint_small"
            # Pass JSON for each row directly
            input = self.input
        }
    }

}
```

IF step in a loop:
```hcl
pipeline "my_steampipe_pipeline" {

    step "instances" {
        base = steampipe.step.query
        sql = "select instance_id, region from aws_ec2_instance"
    }

    step "my_if" {
        for_each = step.instances.output.rows
        base = steampipe.step.if
        if = each.value.region == "us-east-1"
        then {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = each.value
        }
    }

}
```

IF step as a loop filter:
```hcl
pipeline "my_steampipe_pipeline" {

    step "instances" {
        base = steampipe.step.query
        sql = "select instance_id, region from aws_ec2_instance"
    }

    # Simply duplicate the output, but only if the condition passes
    step "my_if" {
        for_each = step.instances.output.rows
        base = steampipe.step.if
        if = each.value.region == "us-east-1"
        then {
            base = steampipe.step.transform
            output = each.value
        }
    }

    # Only called if step.my_if condition matched
    step "my_request" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.my_if.output
    }

}
```

Syntax Option 1 - conditional:
```hcl
pipeline "my_steampipe_pipeline" {

    # Any step can use an if conditional
    step "my_http_with_if" {
        if = self.input.type == "insert"
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.my_filter.output
    }

    # Conditional is a specific if step
    step "my_if" {
        base = steampipe.step.conditional
        if = self.input.type == "insert"
        then {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = step.my_filter.output
        }
    }

    step "my_if_else" {
        base = steampipe.step.conditional
        if = self.input.type == "insert"
        then {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = step.my_filter.output
        }
        else {
            base = steampipe.step.transform
            output = self.input
        }
    }

    step "my_if_elseif_else" {
        base = steampipe.step.conditional
        if = self.input.type == "insert"
        then {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = self.input
        }
        elseif {
            test = self.input.type == "update"
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = self.input
        }
        else {
            base = steampipe.step.transform
            output = self.input
        }
    }

    step "my_case_string" {
        base = steampipe.step.conditional
        case = self.input.type
        when "insert" {
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = self.input
        }
        else {
            base = steampipe.step.transform
            output = self.input
        }
    }

    step "my_case_not_string" {
        base = steampipe.step.conditional
        when { # > 24
            test = self.input.type > 24
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = self.input
        }
        else {
            base = steampipe.step.transform
            output = self.input
        }
    }

    step "my_case_block_string" {
        base = steampipe.step.conditional
        case {
            test = self.input.type
            when "insert" {
                base = steampipe.step.http_request
                url = "https://example.com/with/endpoint"
                # Pass JSON for each row directly
                input = self.input
            }
            when "update" {
                base = steampipe.step.http_request
                url = "https://example.com/with/endpoint"
                # Pass JSON for each row directly
                input = self.input
            }
            else {
                base = steampipe.step.transform
                output = self.input
            }
        }
    }

    step "my_case_block_not_string" {
        base = steampipe.step.conditional
        case {
            test = self.input.type
            when {
                test = test.value > 24
                base = steampipe.step.http_request
                url = "https://example.com/with/endpoint"
                # Pass JSON for each row directly
                input = self.input
            }
            else {
                base = steampipe.step.transform
                output = self.input
            }
        }
    }

}
```

Option 2 - conditional with "test =":
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_if" {
        base = steampipe.step.conditional
        if {
            test = self.input.type == "insert"
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = step.my_filter.output
        }
    }

    step "my_if_else" {
        base = steampipe.step.conditional
        if {
            test = self.input.type == "insert"
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = step.my_filter.output
        }
        else {
            base = steampipe.step.transform
            output = self.input
        }
    }

    step "if_elseif_else" {
        base = steampipe.step.conditional
        if {
            test = self.input.type == "insert"
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = self.input
        }
        elseif {
            test = self.input.type == "update"
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = self.input
        }
        else {
            base = steampipe.step.transform
            output = self.input
        }
    }

    step "my_case_string" {
        base = steampipe.step.conditional
        case {
            test = self.input.type == "insert"
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = self.input
        }
        else {
            base = steampipe.step.transform
            output = self.input
        }
    }

    step "my_case_not_string" {
        base = steampipe.step.conditional
        when { # > 24
            test = self.input.type > 24
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = self.input
        }
        else {
            base = steampipe.step.transform
            output = self.input
        }
    }

}
```

TODO / NEXT STEPS:
* Can if = be used in any step, or only steampipe.step.conditional?
* Finalize the syntax proposal
* Support for_each with conditional? e.g. while or until


#### IF/ELSE/CASE via filter

Notes:
* We need some way to stop processing, it makes no sense to run through all the steps checking each time if the condition was false.

IF via filter:
```hcl
pipeline "my_steampipe_pipeline" {

    # A filter step will only continue further steps if the condition matches.
    step "my_filter" {
        base = steampipe.step.filter
        # HCL style condition
        condition = self.input.type == "insert"
    }

    # Only called if the condition matched
    step "do_for_insert" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.my_filter.output
    }

}
```

IF / ELSE via filter:
```hcl
pipeline "my_steampipe_pipeline" {

    # IF check for insert
    step "insert_filter" {
        base = steampipe.step.filter
        # HCL style condition
        condition = input.type == "insert"
    }

    # Depends on insert_filter, so only called if it matches
    step "do_for_insert" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.insert_filter.output
    }

    # ELSE is done via a matching not condition
    step "not_insert_filter" {
        base = steampipe.step.filter
        # HCL style condition
        condition = input.type != "insert"
    }

    # Depends on insert_filter, so only called if it matches
    step "do_for_not_insert" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.not_insert_filter.output
    }

}
```

CASE via filter:
```hcl
pipeline "my_steampipe_pipeline" {

    # A filter step will only continue further steps if the condition matches.
    step "insert_filter" {
        base = steampipe.step.filter
        # HCL style condition
        condition = input.type == "insert"
    }

    # Depends on insert_filter, so only called if it matches
    step "do_for_insert" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.insert_filter.output
    }

    # Runs in parallel with step.insert_filter
    step "update_filter" {
        base = steampipe.step.filter
        # HCL style condition
        condition = input.type == "update"
    }

    # Depends on update_filter, so only called if it matches
    step "do_for_update" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.update_filter.output
    }

}
```







### Loops

Option 1 (Recommended) - dependency based with explicit explode and implode steps:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
    }

    step "my_query_row" {
        base = steampipe.step.explode
        input = step.my_query.output.rows
    }

    # Option 1 - Each row is exposed as a result of the prior step
    step "my_webhook_loop" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        # input = step.my_query_row.output
        # Or, get a single value out to pass
        body = {
            account_id = step.my_query_row.output.account_id
        }
    }

    # Option 2 - Each result of the prior step returns a different index, used to access
    # inside the full result set.
    step "my_webhook_loop" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Lookup the specific row by index
        body = {
            account_id = step.my_query.output.rows[step.my_query_row.index].account_id
        }
    }

    step "web_webhook_results" {
        base = steampipe.step.implode
        # Other names could be: gather, collect, aggregate, implode
        steps = [
            step.my_webhook_loop
        ]
    }

    step "finished_notification" {
        base = steampipe.step.slack_notification
        # message = "Finished processing ${step.my_webhook_results.output.count} accounts"
        message = "Finished processing ${length(step.my_webhook_results.output.results)} accounts"
    }

}
```

Option 2a - for_each:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
    }

    step "my_webhook_loop" {
        for_each = step.my_query.output.rows

        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        # input = each.value
        # Or, get a single value out to pass
        input = {
            "account_id" = each.value.account_id
        }
    }

    # Option 1
    step "finished_notification" {
        base = steampipe.step.slack_notification
        message = "Finished processing ${step.my_webhook_loop.output.count} accounts"
    }

    # Option 2
    step "finished_notification" {
        base = steampipe.step.slack_notification
        message = "Finished processing ${length(step.my_webhook_loop)} accounts"
    }

}
```

Option 2b - for_each with sub-steps:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
    }

    step "my_webhook_loop" {
        for_each = step.my_query.output.rows

        step "prod" {
            filter {
                where = "tags ->> 'env' = 'prod'"
            }
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = {
                "account_id" = each.value.account_id
            }
        }

        step "other" {
            filter {
                where = "tags ->> 'env' != 'prod'"
            }
            base = steampipe.step.http_request
            url = "https://example.com/with/endpoint"
            # Pass JSON for each row directly
            input = {
                "account_id" = each.value.account_id
            }
        }

    }

    step "finished_notification" {
        base = steampipe.step.slack_notification
        message = "Finished processing ${step.my_webhook_loop.output.count} accounts"
    }

}
```

Option 3 - set flow type (implode, explode, etc) on source:
```hcl
pipeline "my_steampipe_pipeline" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
        flow = "explode"
    }

    step "my_webhook_loop" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        # input = step.my_query_row.output
        # Or, get a single value out to pass
        input = {
            "account_id" = step.my_query_row.output.account_id
        }
        flow = "implode"
    }

    step "finished_notification" {
        base = steampipe.step.slack_notification
        # message = "Finished processing ${step.my_webhook_results.output.count} accounts"
        message = "Finished processing ${length(step.my_webhook_results.output.results)} accounts"
    }

}
```

### Error handling

Simple example:
```hcl
pipeline "main" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
        # Uses the default error handler
    }

    step "my_webhook_loop" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.my_query_row.output
        error = error.http
    }

}

error "http" {
    # Errors are fatal by default, so the pipeline will not be retried. If
    # false, the pipeline should be automatically retried (with a backoff).
    fatal = false
    call = pipeline.slack.error_notify
    text = "Unexpected error calling webhook ${var.user_name}"
}

# There is a default error handler, does not need to be explicitly defined.
# It catches all unhandled error types.
# error "default" {
#   base = pipeline.steampipe.log
#   level = error
#   message = "Unhandled error in ${pipeline.name}"
#   data = error
# }
```

Error handling by status code:
```hcl
pipeline "main" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
        # Uses the default error handler
    }

    step "my_webhook_loop" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Pass JSON for each row directly
        input = step.my_query_row.output

        # timeout set by duration
        timeout = "10s"

        # Option 1 - list of handlers
        error = [
            error.http_429,
            error.http_4xx,
            error.http_5xx,
            error.default
        ]

        # Option 1a - list of handlers, via a local var
        error = local.error_handlers

        # Option 2 - automatically try handlers in defined order
        # Nothing required here

        # Option 3 - Have a defined group of error handlers
        error = error_group.http
    }

}

error "retry_429" {
    filter {
        where = "status_code = 429"
    }
    # TODO - is this max_retries per error type, or per step?
    max_retries = 3
    backoff = "exponential"
    text = "Rate limited, failed after 3 retries"
}

error "timeout" {
    filter {
        where = "status_code = 408"
    }
    max_retries = 3
    backoff = "exponential"
    text = "Timeout, failed after 3 retries"
}

error "fail_4xx" {
    filter {
        where = "status_code >= 400 AND status_code < 500"
    }
    # These are the same, which is better?
    max_retries = 0
    fatal = true
    # User error
    text = "Error: ${error.message}"
}

error "fail_5xx" {
    filter {
        where = "status_code >= 500"
    }
    # These are the same, which is better?
    max_retries = 0
    fatal = true
    # User error
    text = "Internal Error: please report error ${error.id} to support"
}

# There is a default error handler, does not need to be explicitly defined.
# It catches all unhandled error types.
# error "default" {
#   base = pipeline.steampipe.log
#   level = error
#   message = "Unhandled error in ${pipeline.name}"
#   data = error
# }
```


### Logging

Logging and the log level is a property of the service. Logs are emitted from the various pipeline steps and captured according to the log level. Each primitive or function is encouraged to log appropriately for it's own data, similar to how GitHub actions or CI/CD pipelines would create logs.

### Functions

The `function` primitive will automatically detect and compile AWS Lambda compatible functions using Docker for execution by Flowpipe. Our intent is to make it easy to reuse existing code or increase confidence to build for flowpipe knowing your work can be useful later.

Based on the `location` we look in that directory for code of common languages (e.g. Python, Javascript, Golang, Shell, etc). We then choose the appropriate AWS Lambda base image for running it and expect the code to be in that format. The base image causes the function to become available as a URL that we can POST to from our code - we should only start the container for functions when they are actually used by the pipeline, and then leave them "hot" for a period.

We will never be as efficient to run as AWS Lambda - and should not aim to be. Instead, we're trying to be radically easier for development and versatile for scripting.

See the `examples/functions` directory.

#### Function caching

Functions results can be cached:
```hcl
pipeline "my_pipeline_with_default_connection" {
    step "my_func" {
        base = steampipe.step.function
        input = {
            foo = "bar"
        }
        # Like distinct, each cache block is an OR, multiple blocks means AND
        cache {
            ttl = "<duration>"
            until = "<fixed_timestamp>"
            count = 100
        }
    }
}
```

### Calling pipelines

```hcl
pipeline "my_query" {

    param "owner_tag" {
        type = string
    }

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account where tags ->> 'Owner' = ${param.owner_tag}"
    }

    # define a specific pipeline output
    output "rows_output" {
        value = step.my_query.output.rows
    }
}

pipeline "nested_pipelines" {
    step "my_query_pipeline" {
        # This is a step primitive to run a pipeline
        base = steampipe.step.run_pipeline
        # The pipeline to be run, with args passed to the pipeline params
        pipeline = pipeline.my_query
        args = {
            owner_tag = "Nathan"
        }
    }

    step "my_webhook" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # Get the output of the step, which is defined as the pipeline outputs,
        # so in this case just the rows_output field
        input = step.my_query_pipeline.output.rows_output
    }
}
```

#### Control flow

Here is the execution rules for pipelines:
1. Run everything in parallel if possible.
2. Determine data dependencies between steps, and build an execution order from that.
3. A step can be skipped by setting `skip = true` on the step. (Question - should dependent steps be automatically skipped?)
4. Parallelism can be controlled on both a pipeline and a step level. On the pipeline, it limits the number of steps (of any or the same type) being run concurrently. On the step, it limits the number of instances of that single step that are running concurrently. 

Option 6 for_each - parallel pipelines
```hcl
pipeline "main" {

    # Get all the rows
    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title, tags from aws_account"
        # Uses the default error handler
    }

    # For each row, run a pipeline
    step "prod" {
        for_each = step.my_query.output.rows
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_prod
        args = each.value
    }

    # For each row, run another pipeline
    step "prod_parallel" {
        for_each = step.my_query.output.rows
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_prod_2
        args = each.value
    }

}
```

Option 6 for_each - chaining one pipeline and then another
```hcl
pipeline "main" {

    # Get all the rows
    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title, tags from aws_account"
        # Uses the default error handler
    }

    # For each row, run a pipeline
    step "prod" {
        for_each = step.my_query.output.rows
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_prod
        args = each.value
    }

    # For each pipeline run, run another pipeline
    step "prod_2" {
        for_each = step.prod
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_prod_2
        args = each.value.output
    }

}
```


Option 5 (Recommended, dependency based & simple to read and follow) - Filter them out
```hcl
pipeline "main" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title, tags from aws_account"
        # Uses the default error handler
    }

    step "my_query_row" {
        base = steampipe.step.explode
        input = step.my_query.output.rows
        target = "row"
    }

    step "prod_row" {
        base = steampipe.step.filter
        input = step.my_query_row.output.row
        filter {
            where = "tags ->> 'env' = 'prod'"
        }
    }

    step "prod" {
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_prod
        args = step.prod_row
    }

    step "other_row" {
        base = steampipe.step.filter
        input = step.my_query_row.output.row
        filter {
            where = "tags ->> 'env' != 'prod'"
        }
    }

    step "other" {
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_prod
        args = step.other_row
    }

    step "wait_then_finalize" {
        base = steampipe.step.implode
        steps = [
            step.prod,
            step.other
        ]
    }

}
```

Option 1 - Choice step, calling other pipelines (not steps) - messes with the dependency model
```hcl
pipeline "main" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title, tags from aws_account"
        # Uses the default error handler
    }

    step "my_query_row" {
        base = steampipe.step.explode
        input = step.my_query.output.rows
    }

    step "choice" {
        input = step.my_query_row.output.value
        # input = step.my_query.output.rows[step.my_query_row.output.index]
        case {
            where = "tags ->> 'env' = 'prod'"
            # TODO - how to call with args?
            next = [ pipeline.main_prod ]
        }
        default {
            next = [ pipeline.main_dev ]
        }
    }

}
```

Option 2 - Dependency flow with where condition on steps
```hcl
pipeline "main" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title, tags from aws_account"
        # Uses the default error handler
    }

    step "my_query_row" {
        base = steampipe.step.explode
        input = step.my_query.output.rows
    }

    step "prod" {
        # Decide if the step should run...
        input = step.my_query_row.output.value
        where = "tags ->> 'env' = 'prod'"
        # Step definition
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_prod
        args = input
    }

    step "other" {
        # Decide if the step should run...
        input = step.my_query_row.output.value
        where = "tags ->> 'env' != 'prod'"
        # Step definition
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_other
        args = input
    }

}
```

Option 3 - Expressions in depends on
```hcl
pipeline "main" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title, tags from aws_account"
        # Uses the default error handler
    }

    step "my_query_row" {
        base = steampipe.step.explode
        input = step.my_query.output.rows
        target = "row"
    }

    step "prod" {
        # Option 1 - allow HCL mixed in SQL
        depends_on = [ "step.my_query_row.row.tags['env'] = 'prod'" ]
        # Option 2 - try to separate HCL from SQL
        depends_on = [ "${step.my_query_row.row}.tags ->> 'env' = 'prod'" ]
        # Option 3 - separate the parts (implicit depends_on)
        input = step.my_query_row.row
        where = "tags ->> 'env' = 'prod'" # assumes data in row
        # Option 4 - like 3, but in a block
        depends_on {
            input = step.my_query_row.row
            where = "tags ->> 'env' = 'prod'"
        }
        # Step definition
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_prod
        args = input
    }

    step "other" {
        depends_on = [ "${step.my_query_row.row}.tags ->> 'env' != 'prod'" ]
        # Step definition
        base = steampipe.step.run_pipeline
        pipeline = pipeline.main_prod
        args = input
    }

    step "wait_then_finalize" {
        depends_on = [step.prod, step.other]
        base = steampipe.step.implode
        steps = [
            step.prod,
            step.other
        ]
    }

}
```


#### Filtering

TODO - How to do an IF (filter)?
TODO - How to do an IF/ELSE?
TODO - How to do a CASE?

Filtering on a single object before doing more steps:
```hcl
pipeline "main" {
    step "my_http_request" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # method = "post"
        # content_type = "application/json"
        body = {
            my_input = "my_value"
        }
    }

    # Option 1 - SQL syntax
    step "my_filter" {
        base = steampipe.step.filter
        sql = "select * from ${step.my_http_request} where output -> 'foo' = 'bar'"
    }

    # Option 2 - HCL syntax
    step "my_filter" {
        base = steampipe.step.filter
        filter = step.my_http_request.output.foo == 'bar'
    }

    # Option 3 - filter on input
    step "my_filter" {
        base = steampipe.step.filter
        input = step.my_http_request.output
        filter {
            where = "foo = 'bar'"
        }
    }

    step "my_http_reaction" {
        base = steampipe.step.http_request
        url = "https://example.com/with/reaction"
        # method = "post"
        # content_type = "application/json"
        body = {
            my_input = each.value.foo
        }
    }

    ...
}
```

Filtering on a single object using count syntax:
```hcl
pipeline "main" {
    step "my_http_request" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # method = "post"
        # content_type = "application/json"
        body = {
            my_input = "my_value"
        }
    }

    step "my_http_reaction" {
        count = step.my_http_request.output.foo == 'bar' ? 1 : 0
        base = steampipe.step.http_request
        url = "https://example.com/with/reaction"
        # method = "post"
        # content_type = "application/json"
        body = {
            my_input = step.my_http_request.output.foo
        }
    }

    ...
}
```

Filtering on rows before doing more steps:
```hcl
pipeline "main" {

    step "my_query" {
        base = steampipe.step.query
        query = "select account_id, title, tags from aws_account"
    }

    step "my_http_reaction" {
        for_each = [for v in step.my_query_output.rows: v if v.tags.env != 'prod']
        base = steampipe.step.http_request
        url = "https://example.com/with/reaction"
        # method = "post"
        # content_type = "application/json"
        body = {
            my_input = each.value.foo
        }
    }

    ...
}
```


#### Deduplication using filter

Dedup on the main event, no actual explode
```hcl
pipeline "main" {
    step "my_http_request" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # method = "post"
        # content_type = "application/json"
        body = {
            "my_input" = "my_value"
        }
    }

    step "my_dedup" {
        base = steampipe.step.filter
        input = step.my_http_request.output
        distinct {
            on = "account_id"
        }
    }

    ...
}
```

Dedup on a window of time AND event count:
```hcl
pipeline "main" {
    step "my_http_request" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # method = "post"
        # content_type = "application/json"
        body = {
            "my_input" = "my_value"
        }
    }

    step "my_dedup" {
        base = steampipe.step.filter
        input = step.my_http_request.output
        # Multiple distinct blocks means ALL of
        # TODO - should it be the other way around?
        distinct {
            since = "5m" # a duration
            count = 100
        }
    }

    ...
}
```

Dedup on a window of time OR event count:
```hcl
pipeline "main" {
    step "my_http_request" {
        base = steampipe.step.http_request
        url = "https://example.com/with/endpoint"
        # method = "post"
        # content_type = "application/json"
        body = {
            "my_input" = "my_value"
        }
    }

    step "my_dedup" {
        base = steampipe.step.filter
        input = step.my_http_request.output
        # Multiple distinct blocks means ANY of
        # TODO - should it be the other way around?
        distinct {
            since = "5m" # a duration
        }
        distinct {
            count = 100
        }
    }

    ...
}
```

#### Transpose

Transform data through a step:
```hcl
pipeline "my_transpose" {

    step "my_query" {
        base = steampipe.step.query
        sql = "select account_id, title from aws_account"
        # Uses the default error handler
    }

    # Option 1 - SQL, like a WITH
    step "my_transpose" {
        base = steampipe.step.query
        sql = "select account_id as aws_account_id from ${step.my_query}"
    }

    # Option 1 - jq
    step "my_transpose" {
        base = steampipe.step.transpose
        input = step.my_query.output.rows
        filter = ".[] | [{"aws_account_id":.account_id}]"
    }
}
```


### Providers & credentials

#### Choosing the connection for a step

Hard-coded connection for a step:
```hcl
pipeline "my_pipeline_with_default_connection" {
    step "stop_instance" {
        base = steampipe.step.function
        input = {
            tags = {
                env = "prod"
            }
        }
        # Connection to use for the step
        connection = "aws_01"
    }
}
```

Default connection for a step. This is the first connection in the search_path:
```hcl
pipeline "my_pipeline_with_default_connection" {
    step "stop_instance" {
        base = steampipe.step.function
        input = {
            tags = {
                env = "prod"
            }
        }
    }
}
```

To target a connection by plugin type, you can specify the plugin type:
```hcl
pipeline "my_pipeline_with_default_connection" {
    step "stop_instance" {
        base = steampipe.step.function
        input = {
            tags = {
                env = "prod"
            }
        }
    }
    # Cause the connection to be the first in search path for this plugin type
    plugin = "aws"
}
```

```
pipeline {
    param "github_credentials" {
        default = my.default.github.connection
    }
}

step "container" "foo" {
    image = "amazon/aws-cli"
    cmd = [...]
}

step "container" "foo" {
    image = "github/github-cli"
    cmd = [...]
    credentials = "my_github"
    credentials_family = "github"
}

step "container" "foo" {
    image = "slack/slack-cli"
    cmd = [...]
}

step "container" "foo" {
    image = "azure/azure-cli"
    cmd = [...]
}
```


```
step "container" "foo" {

    # hardcoded connection
    creds = "aws_01"

    # dynamic connection
    creds = param.connection

    # default connection uses first in search path
    # nothing

    # get first connection (not aggregator) from the search of the plugin type
    plugin = "aws"
}

Connections can be set dynamically from input data:
```hcl
pipeline "my_pipeline_with_default_connection" {
    step "stop_instance" {
        base = steampipe.step.function
        input = {
            tags = {
                env = "prod"
            }
        }
    }
    connection = input.connection
}
```

#### How are credentials calculated / made available?

Plugins implement a standard interface to return credentials and other common information. The exact format will vary by plugin (e.g. AWS has an access key, secret key and session token while other services might just have a token).

For example, when running a function for an AWS connection it would make the following environment variables available: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_REGION`.



## Developer Experience

### CLI commands

```shell
# Manually run a pipeline
$ flowpipe run pipeline.my_webhook_pipeline

# Manually run a pipeline with input
$ flowpipe run pipeline.my_webhook_pipeline --input '{"my_input": "my_value"}'

# Pipe standard input to the pipeline
$ echo '{"my_input": "my_value"}' | flowpipe run pipeline.my_webhook_pipeline

# Run flowpipe as service
$ flowpipe service start
[flowpipe] Starting service
[flowpipe] Loading mod
[flowpipe] Loading pipelines
[flowpipe] Loading triggers
[flowpipe] Starting triggers
[flowpipe] Starting trigger webhook.my_webhook
[flowpipe] Starting trigger query.my_query
[flowpipe] Starting trigger cron.my_cron
[flowpipe] Started service
[🪝.my_webhook       ] Triggered {"foo": "bar"}
[🕑.my_cron          ] Triggered {"time": "2023-02-07T0915:33"}
[🕑.my_cron          ] Running pipeline.my_scheduled_pipeline
[╦.my_scheduled_pipe…] Queued
[╦.my_scheduled_pipe…] Queued
[╦.my_scheduled_pipe…] Queued
```

IDEA - CLI should be more of a console with areas. It could list the triggers that are listening with basic stats on their number of requests, then have an area for user prompts / inputs and finally an area of log output to see activity.

### SDK

Flowpipe runs as a service with an API, which is used by both the Flowpipe CLI
and various SDKs to directly manage running pipelines etc. This is similar to
Vault, Docker, Kubernetes, etc.

Through the SDK, Flowpipe can act as a full microservice coordination engine,
managing functions and activities from key points in your own code / app.

For example:
```go
client := flowpipe.NewClient("http://localhost:8080")
pipelineID, err := client.RunPipeline("my_pipeline", flowpipe.RunPipelineInput{
    Input: map[string]interface{}{
        "my_input": "my_value",
    },
})
if err != nil {
    panic(err)
}
client.WaitForPipeline(pipelineID)
```

## Business model

Flowpipe will be released under the AGPLv3 license.

We'll provide a cloud hosted offering, with simple upload of code via GitHub etc. We'll charge for execution time, and also per-user for workflow approvals etc.

In the future, we'll also provide a self-hosted option, with a simple binary to run. We'll charge for support and maintenance.

### Buyer based tiers

* Developer [CLI] - Free, open source CLI.
* Developer [Cloud] - Free, limited hosting service.
* Team - centralized hosting, private sharing, work with github orgs, multi-user.
* Enterprise - SAML, dedicated instance, etc.

### Positioning

Simple replacement for step functions and cloud scripting. Use it as a SOAR or for workflows from your Vercel website. (Similar to how PlanetScale is a simple replacement for RDS.)

Not a:
* Replacement for Temporal
* Replacement for Zapier
* Replacement for large-scale Lambda
* Replacement for CI/CD

When to switch:
* Temporal too complex to operate and hard to understand
* Zapier limited, ready to customize
* Lambda + Step Functions is too much dev and hard to fathom for a simple task
* GitHub actions running as a cron and not using GitHub events
* Python scripts need workflow approvals and coordination

### Flowpipe vs Competitors

Flowpipe is:
* Code-forward (HCL + SQL)
* Lambda-compatible for functions
* A single binary, easy to operate and scale
* Designed for developers

Flowpipe is not:
* As fast or scalable as Lambda functions
* A no-code tool for non-developers
* Storing state in workflows

#### Flowpipe vs Temporal

Temporal is a workflow engine that believes in mixing the workflow definitions into application code. Workflows are stateful and long-running. It's a great tool for complex scenarios where the complexity to operate is worth the benefits.

Flowpipe uses a low-code approach to pipeline definitions in HCL. Common tasks (e.g. web requests, notifications, etc) work out of the box, while Lambda-compatible functions can be used for custom code. A single binary, it's easy to operate and scale both as a developer or in production.

#### Flowpipe vs Zapier

Zapier is a no-code workflow tool intended for non-technical users. It's a great tool for simple workflows, but difficult to manage, control and extend for more complex scenarios.

Flowpipe has a code-first approach, giving control over versioning and operation. It can be used for simple workflows without custom code, but also extends to more complex and large-scale scenarios.

#### Flowpipe vs AWS Lambda & Step Functions

AWS Lambda with Step Functions allows developers to create workflows with custom code. It requires significant experitise to create and deploy, but offers large scale at low cost.

Flowpipe is a low-code tool to create workflows and pipelines. It is compatible with AWS Lambda functions, but allows developers to automatically build and test them locally. It's significantly easier for simple cases, and a great way to build more complex coordination flows. It can be run anywhere and scales to the level required by all but the largest apps.



## How it works

A mod defines a collection of pipelines and triggers. The mod may be
loaded to then start listening for events and executing pipelines.

## Event Sourcing and CQRS

Here is the sequence when starting:
* Queue - waiting to load the mod
* Load - Load the full mod definition including pipelines and triggers
* Start - Start execution of the mod
* Plan - Plan the next tasks to be executed for the mod

The mod can end in a few ways:
* Finish - Everything is done, clean shutdown
* Fail - Something went wrong, forced shutdown

When running a pipeline (regardless of the trigger), the sequence is:
* Queue - waiting to start the pipeline
* Load - Load the pipeline definition
* Start - Start execution of the pipeline
* Plan - Plan the next steps to be executed for the pipeline
* Queue Step - Queue the step to be executed
* Load Step - Load the step definition
* Start Step - Start execution of the step
* Execute Step - Execute the step
* Finish Step - Finish execution of the step
* Fail Step - Fail execution of the step due to an error


## Runtime identifiers

The mod is running, waiting for triggers.
Each trigger starts a pipeline, which has a unique ID.
Each step in the pipeline has a unique ID.

The IDs above are nested, giving a StackID.


## Questions

* What is the max event size we support?
* Can we have an object store to reuse for data?
* Can triggers be synchronous in response?




# BACKUP / OLD

## Choice - DEPRECATED, see control flow section above

Choice steps are used to direct the workflow based on current data or values. The data is tested against a series of rules. The first rule to match will execute it's next steps and stop further processing.

Example Choice - split processing by tag:
```hcl
step "my_choice" {
    base = steampipe.step.choice

    rule {
        where = "tags ->> 'env' = 'prod'"
        next = [ "prod_step" ]
    }

    rule {
        where = "tags ->> 'env' = 'dev'"
        next = [ "dev_step" ]
    }

    # Default is to do nothing if no rule is matched.
    # Can also be defined explicitly as:
    # rule {
    #   // Not matched, do nothing
    #   next = []
    # }

}
```

Example Choice - dynamic:
```hcl
step "my_choice" {
    base = steampipe.step.choice

    rule {
        # Where clause run against input data
        where = "tags ->> 'env' in ('dev', 'prod')"
        # Use jq for dynamic next step calculation?
        next = [ "{{tags ->> 'env'}}_step" ]
        # Use HCL for dynamic next step calculation?
        next = [ "${input.tags["env"]}_step" ]
    }

}
```

Filter Example - only continue processing for buckets with a name like foo-%:
```hcl
step "my_choice" {
  base = steampipe.step.choice

  rule {
    // If matched, then continue to next step
    match = "name like 'foo-%'"
  }

}
```

Filter Example - stop processing for all buckets where the name is not like foo*
```hcl
step "my_choice" {
  base = steampipe.step.choice

  rule {
    // If matched, then stop processing
    match = "name like 'foo-%'"
    next = []
  }

  rule {
    // * means continue processing with next matched step
    next = [ "*" ]
  }

}
```

Example - remove duplicates:
```hcl
pipeline "my_dedup" {

    # PRE: Triggered by a webhook of events for example

    # Option 1 (recommended) - enhancement to choice
    step "dedup" {
        base = steampipe.step.choice
        match = "not duplicate(event_type, event_name)"
    }

    # Option 2 - specific dedup function
    step "dedup" {
        base = steampipe.step.deduplicate
        select = "event_type, event_name"
    }

}
```


## Parallel execution flow

START:
* Find all steps with no pre-reqs
* Start them running

STEP FINISHED:
* Add it to the list of completed steps (with their data)
* Find all steps depending on the completed step
* If all pre-reqs are now met for the step, start it
* Otherwise, if there are no more running steps, then the pipeline has finished


## Timeouts, retries and backoff

Basic principles:
* When a step runs, it can result in a success (with output) or an error.
* A timeout of the step is an error.
* Errors can be: retryable, ignorable (keep running), end of the thread (abort at this step, but run other threads), or fatal for the whole pipeline, 
* When retrying, there are options for how many times to retry and how long to wait between attempts.

Every step supports configuration of the above. For example:
```hcl
step "my_query" {
    base = steampipe.step.query
    sql = "select account_id, title from aws_account"
    flow {
        timeout = "10s"
        max_retries = 5
        backoff = "exponential"
        backoff_base = "2s"
        fatal_where = [ "status_code in (304, 400, 401, 403)" ]
    }
}
```

TODO - Should we have shared retry / error handler blocks?
TODO - What's the default?
TODO - Multiple *_where lines / blocks to match different types?