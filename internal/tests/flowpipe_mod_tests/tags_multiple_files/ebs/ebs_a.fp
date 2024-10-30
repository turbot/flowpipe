locals {
    foo = "bar"
}

trigger "query" "detect_and_correct_ebs_snapshots_exceeding_max_age" {
  title         = "Detect & Correct EBS Snapshots Exceeding Max Age"
  description   = "Detects EBS snapshots exceeding max age and runs your chosen action."
  // documentation = file("./ebs/docs/detect_and_correct_ebs_snapshots_exceeding_max_age_trigger.md")
  tags          = merge(local.aws_thrifty_common_tags, { class = "unused" })

  schedule = "daily"
  database = "foo"
  sql      = "bar"

  capture "insert" {
    pipeline = pipeline.merging_tags
    args = {
      items = self.inserted_rows
    }
  }
}

pipeline "merging_tags" {
    tags = merge(local.ebs_common_tags, { class = "unused" })

    step "transform" "name" {
        value    = "hello"
    }
}