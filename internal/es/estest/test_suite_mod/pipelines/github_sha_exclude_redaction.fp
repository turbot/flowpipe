pipeline "github_sha_exclude_redaction" {

  step "transform" "test_sha" {
    value = "https://github.com/turbot/flowpipe/commit/7a2c8fd9789a9b6rc8f29c41b42036823e2fceab"
  }

  # The SHA in the output should not be redacted
  output "sha" {
    value = "https://github.com/turbot/flowpipe/commit/7a2c8fd9789a9b6rc8f29c41b42036823e2fceab"
  }
}