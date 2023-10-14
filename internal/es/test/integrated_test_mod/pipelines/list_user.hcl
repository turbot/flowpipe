pipeline "list_users" {
  description = "List the users."

  param "password" {
    type        = string
    description = "The API token for authorization."
    default     = "Sukla@123"
  }

  param "user_email" {
    type        = string
    description = "The user email ID of the user the account belongs to."
    default     = "madhushree@turbot.com"
  }

  step "http" "list_users" {
    title  = "List users"
    method = "get"
    url    = "https://turbotsupport.zendesk.com/api/v2/users.json"

    basic_auth  {
      username = param.user_email
      password = param.password
    }

    request_headers = {
      Content-Type  = "application/json"
      //Authorization = "Basic ${base64encode("${param.user_email}:${param.password}")}"
    }
  }

  output "users" {
    description = "The list of users associated to the account."
    value       = jsondecode(step.http.list_users.response_body).users
  }
}