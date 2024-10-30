pipeline "param_various" {

  param "foo" {
    default = "bar"
  }

  param "list_of_string" {
    type = list(string)
    default = ["foo", "bar"]
  }

  param "map_of_number" {
    type = map(number)
    default = {
          "foo": 1
          "bar": 2
    }
  }

  param "map_of_bool" {
    type = map(bool)
    default = {
      "foo": true
      "bar": false
    }
  }

  param "map_of_list_of_number" {
    type = map(list(number))
    default = {
      "foo": [1, 2]
      "bar": [3, 4]
    }
  }

  param "map_of_a_map_of_a_bool" {
    type = map(map(bool))
    default = {
      "foo": {
        "bar": true
        "baz": false
      }
      "qux": {
        "quux": true
        "corge": false
      }
    }
  }

  param "map_of_any" {
    type = map(any)
    default = {
      "foo": "bar"
      "baz": 42
      "qux": true
    }
  }

  param "just_map" {
    type = map
    default = {
      "foo": "bar"
      "baz": "qux"
    }
  }

  param "list_of_list_of_string" {
    type = list(list(string))
    default = [["foo", "bar"], ["baz", "qux"]]
  }

  param "list_of_map_of_bool" {
    type = list(map(bool))
    default = [
      {
        "foo": true
        "bar": false
      },
      {
        "baz": true
        "qux": false
      }
    ]
  }

  param "list_of_map_of_list_of_number" {
    type = list(map(list(number)))
    default = [
      {
        "foo": [1, 2]
        "bar": [3, 4, 5, 6]
      },
      {
        "baz": [5, 6]
        "qux": [7, 8]
      }
    ]
  }
  param "list_of_map_of_list_of_string" {
    type = list(map(list(string)))
    default = [
      {
        "foo": ["123", "123"]
        "bar": ["a"]
      },
      {
        "baz": ["a", "b", "c"]
        "qux": ["a"]
      }
    ]
  }  
}
