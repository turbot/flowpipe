pipeline "empty_slice" {
    param "empty_input_string" {
        type = list(string)
        default = []
    }

    param "empty_input_number" {
        type = list(number)
        default = []
    }

    step "transform" "empty_list" {
        value = []
    }

    step "transform" "empty_list_string" {
        value = param.empty_input_string
    }

    output "val" {
        value = step.transform.empty_list.value
    }

    output "empty_output" {
        value = []
    }

    output "empty_input_number" {
      value = param.empty_input_number
    }
}

pipeline "empty_slice_http" {
    step "http" "empty_list" {
        url = "http://localhost:7104/empty_array.json"
    }

    output "val" {
        value = step.http.empty_list.response_body
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