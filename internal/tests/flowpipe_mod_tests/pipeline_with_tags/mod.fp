mod "test_mod" {
  title = "my_mod"
}



locals {
  aws_thrifty_common_tags = {
    category = "Cost"
    plugin   = "aws"
    service  = "AWS"
  }
}

locals {
  ebs_common_tags = merge(local.aws_thrifty_common_tags, {
    service = "AWS/EBS"
  })
}

locals {
  s3_common_tags = merge(local.aws_thrifty_common_tags, {
    service = "AWS/S3"
  })
}


pipeline "simple_tags" {
    tags = {
        Foo = "Bar"
        Baz = "Qux"
    }

    step "transform" "name" {
        value    = "hello"
    }
}

pipeline "merging_tags" {
    tags = merge(local.ebs_common_tags, { class = "unused" })

    step "transform" "name" {
        value    = "hello"
    }
}

trigger "schedule" "every_hour_trigger_on_if" {
    description = "trigger that will run every hour"
    schedule    = "hourly"

    documentation = file("./docs/one.md")

    tags = merge(local.ebs_common_tags, { class = "unused" })

    pipeline    = pipeline.simple_tags
}
