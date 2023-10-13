pipeline "for_each_http_url" {

  param "http_url" {
    type = list(string)
    default = ["https://jsonplaceholder.typicode.com/posts", "http://api.open-notify.org/astros.jsons", "http://api.open-notify.org/astros.json"]
  }

  step "http" "for_each_url_step" {
    for_each = param.http_url
    url = each.value
  }
}

pipeline "for_each_http_url_map" {

  param "http_url" {
    type = map(bool)
    default = {
            "https://jsonplaceholder.typicode.com/posts" = true
            "http://api.open-notify.org/astros.jsons" = true
            "http://api.open-notify.org/astros.json" = true
        }
  }

  step "http" "for_each_url_step" {
    for_each = param.http_url
    url = each.key
  }
}