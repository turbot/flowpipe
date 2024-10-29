connection "gcp_1" {
  plugin      = "gcp"
  project     = "project-aaa"
  credentials = "/home/me/my-service-account-creds-for-project-aaa.json"
}

connection "gcp_2" {
  plugin      = "gcp"
  project     = "project-bbb"
  credentials = "/home/me/my-service-account-creds-for-project-bbb.json"
}