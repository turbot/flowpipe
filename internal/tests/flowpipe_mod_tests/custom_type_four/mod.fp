mod "custom_type_four" {
}

# set from fpvars file
variable "conn" {
    type = connection.steampipe
}

variable "list_of_conns" {
    type = list(connection)
    default = [
        connection.aws.example,
        connection.aws.example_2,
        connection.aws.example_3
    ]
}

variable "conn_generic" {
    type = connection
      default = connection.aws.example

}

variable "list_of_conns_generic" {
    type = list(connection)
     default = [
            connection.aws.example,
            connection.aws.example_2,
            connection.aws.example_3
        ]
}


pipeline "database_connection_ref" {
     step "query" "select" {
        database = connection.steampipe.default
        sql = "SELECT 1"
    }
}

pipeline "database_var_connection_ref" {
     step "query" "select" {
        database = var.conn
        sql = "SELECT 1"
    }
     step "transform" "echo" {
        value = step.query.select
    }
     output "val" {
        value = step.transform.echo.value
    }

}

pipeline "database_param_var_connection_ref" {
    param "conn" {
        type = connection
        default =  var.conn
    }
     step "query" "select" {
        database = param.conn
        sql = "SELECT 1"
    }
    step "transform" "echo" {
       value = step.query.select
   }
    output "val" {
       value = step.transform.echo.value
   }
}

pipeline "database_var_idx_connection_ref" {
     step "query" "select" {
        database = var.list_of_conns_generic[0]
        sql = "SELECT 1"
    }
    step "transform" "echo" {
       value = step.query.select
   }
    output "val" {
       value = step.transform.echo.value
   }
}

pipeline "database_param_string" {
     step "query" "select" {
        database = "postgres://steampipe@127.0.0.1:9193/steampipe"
        sql = "SELECT 1"
    }
     step "transform" "echo" {
       value = step.query.select
    }
    output "val" {
       value = step.transform.echo.value
    }
}
