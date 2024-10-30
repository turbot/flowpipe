connection "datadog_1" {
  plugin = "datadog"

  api_key = "1a2345bc6d78e9d98fa7bcd6e5ef56a7"
  api_url = "https://api.datadoghq.com/"
  app_key = "b1cf234c0ed4c567890b524a3b42f1bd91c111a1"
}

connection "datadog_2" {
  plugin = "datadog"

  api_key = "1a2345bc6d78e9d98fa7bcd6e5ef57b8"
  api_url = "https://api.datadoghq.com/"
  app_key = "b1cf234c0ed4c567890b524a3b42f1bd91c222b2"
}