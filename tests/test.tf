terraform {
  required_providers {
    scalardb = {
      source  = "scalar-labs/scalardb"
      version = "0.1.0"
    }
  }
}

provider "scalardb" {
  host     = var.scalardb_host
  port     = var.scalardb_port
  username = var.scalardb_username
  password = var.scalardb_password
}

variable "scalardb_host" {
  description = "The host address of the ScalarDB server."
  type        = string
  default     = "localhost"
}

variable "scalardb_port" {
  description = "The port of the ScalarDB server."
  type        = number
  default     = 60051
}

variable "scalardb_username" {
  description = "Username for ScalarDB authentication."
  type        = string
  default     = "admin"
}

variable "scalardb_password" {
  description = "Password for ScalarDB authentication."
  type        = string
  default     = "password"
  sensitive   = true
}

resource "scalardb_namespace" "test" {
  name               = "test_namespace"
  replication_factor = 3
  strategy_class     = "SimpleStrategy"
  durable_writes     = true
}

resource "scalardb_table" "users" {
  namespace      = scalardb_namespace.test.name
  name           = "users"
  partition_key  = ["user_id"]
  clustering_key = ["created_at"]

  column {
    name = "user_id"
    type = "TEXT"
  }

  column {
    name = "created_at"
    type = "BIGINT"
  }

  column {
    name = "name"
    type = "TEXT"
  }

  column {
    name = "email"
    type = "TEXT"
  }

  column {
    name = "age"
    type = "INT"
  }

  compaction_strategy = "SizeTieredCompactionStrategy"

  clustering_order = {
    created_at = "DESC"
  }
}

resource "scalardb_table" "posts" {
  namespace      = scalardb_namespace.test.name
  name           = "posts"
  partition_key  = ["user_id"]
  clustering_key = ["post_id"]

  column {
    name = "user_id"
    type = "TEXT"
  }

  column {
    name = "post_id"
    type = "TEXT"
  }

  column {
    name = "title"
    type = "TEXT"
  }

  column {
    name = "content"
    type = "TEXT"
  }

  column {
    name = "created_at"
    type = "BIGINT"
  }
}

output "namespace_name" {
  value = scalardb_namespace.test.name
}

output "users_table_id" {
  value = scalardb_table.users.id
}

output "posts_table_id" {
  value = scalardb_table.posts.id
}
