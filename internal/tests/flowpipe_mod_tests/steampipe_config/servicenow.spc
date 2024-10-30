connection "servicenow_1" {
  plugin = "servicenow"

  instance_url = "https://test.service-now.com"
  username     = "flowpipe"
  password     = "somepassword"
}

connection "servicenow_2" {
  plugin = "servicenow"

  instance_url = "https://test1.service-now.com"
  username     = "flowpipe"
  password     = "somepassword1"
}