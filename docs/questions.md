
In this error block:

```yaml
pipeline "subscribe" {
  param "subscriber" {
    type = "string"
  }

  step "http" "my_request" {
    url           = "https://myapi.local/subscribe"
    method        = "post"
    body          = jsonencode({
      name = param.subscriber
    })

    error {
      if      = attempt.status_code == 429
      retries = 3
    }
  }
}
```
how do I define another retry? Say if the status code == 500 retry =2

`john`

in the present design, there is only one error block and this is not possible.  Do you think this likely to be a requirement?


`victor`

Not sure .. just wondering how to do that if I want to
New


`john`

actually, you could do this specific example with conditionals i think:
pipeline "subscribe" {
  param "subscriber" {
    type = "string"
  }

  step "http" "my_request" {
    url           = "https://myapi.local/subscribe"
    method        = "post"
    body          = jsonencode({
      name = param.subscriber
    })

    error {
      if      = attempt.status_code == 429 || attempt.status_code == 500
      retries = attempt.status_code == 429 ? 3 : 2
    }
  }
}


---

Another question. Consider this step:

```yaml
steps:
  sleep_1:
    type: "sleep"
    name: "sleep_1"
    depends_on: []
    for: '["1s", "2s", "150ms", "300ms", "450ms", "600ms"]'
    input: '{"duration": "{{.each.value}}"}'
```
Each of the for is a separate execution. Say I have error retry setting, if I retry the step the entire step is retries yes? So we will execute every single for again?

`john`

a for_each creates multiple step instances, and each instance runs/errors/retries independently

---

Say a step has 3 retries, each of them has different errors. Which error should we use in the pipeline? The last one?
