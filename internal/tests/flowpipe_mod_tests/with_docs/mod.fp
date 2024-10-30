
mod "test_mod" {
  title = "my_mod"

  tags = {
    foo = "bar"
    green = "day"
  }
}

pipeline "doc_from_file" {
  description = "inline doc"
  documentation = file("./docs/one.md")
}


trigger "query" "t" {
  title = "Trigger Title"
  description = "Trigger Description"
  documentation = file("./docs/two.md")
  enabled  = false
  schedule = "0 0 * * *"
  database = ""
  sql      = ""

  capture "insert" {
    pipeline = pipeline.p
    args = {
      items = self.inserted_rows
    }
  }
}

pipeline "p" {
  step "transform" "t" {
    value = "hi"
  }
}