pipeline "simple_http_loop" {

  step "http" "list_workspaces" {
    url    = "https://latestpipe.turbot.io/api/v1/org/latesttank/workspace/?limit=3"
    method = "get"

    request_headers = {
      Content-Type  = "application/json"
      Authorization = "Bearer ${param.pipes_token}"
    }

    loop {
      until = result.response_body.next_token != null
      
      url   = "https://latestpipe.turbot.io/api/v1/org/changed_url/workspace/?limit=3&next_token=${result.response_body.next_token}"
    }
  }
}
