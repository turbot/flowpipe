pipeline "text_expr" {
  step "transform" "text_1" {
    value = "foo"
  }

  step "transform" "text_2" {
    value = "bar ${step.transform.text_1.value} baz"
  }

  step "transform" "text_3" {
    value = "bar ${step.transform.text_2.value} baz ${step.transform.text_1.value}"
  }
}

pipeline "expr_func" {
  step "transform" "text_title" {
    value = title("Hello World")
  }
}

pipeline "expr_within_text" {
  step "transform" "text_title" {
    value = "Hello ${title("world")}"
  }
}


pipeline "expr_depend_and_function" {
  step "transform" "text_1" {
    value = "foo"
  }

  step "transform" "text_1_a" {
    value = title("foo")
  }


  step "transform" "text_2" {
    value = title("bar ${step.transform.text_1.value} baz")
  }

  step "transform" "text_3" {
    value = "output2 ${title(step.transform.text_2.value)} func(output1) ${title(step.transform.text_1.value)}"
  }
}
