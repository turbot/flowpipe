pipeline "empty_slice" {

    step "transform" "empty_list" {
        value = []
    }

    output "val" {
        value = step.transform.empty_list.value
    }
}


pipeline "list_contact_lists" {
  title       = "List Contact Lists"
  description = "Returns an array of all of your contact lists."

  param "api_key" {
    type        = string
    default     = "SG..."
  }

  step "http" "list_contact_lists" {
    method = "get"
    url    = "https://api.sendgrid.com/v3/marketing/lists?page_size=1000"

    request_headers = {
      Content-Type  = "application/json"
      Authorization = "Bearer ${param.api_key}"
    }
  }

  output "total_lists" {
    description = "Array of all contact lists."
    value       = step.http.list_contact_lists.response_body
  }
}