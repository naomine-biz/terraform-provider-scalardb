terraform {
  required_providers {
    scalardb = {
      source  = "scalar-labs/scalardb"
      version = "0.1.0"
    }
  }
}

provider "scalardb" {
  host     = "localhost"
  port     = 60051
  username = "admin"
  password = "password"
}

resource "scalardb_namespace" "example" {
  name               = "example_namespace"
  replication_factor = 3
  strategy_class     = "SimpleStrategy"
  durable_writes     = true
}

resource "scalardb_table" "users" {
  namespace      = scalardb_namespace.example.name
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
  namespace      = scalardb_namespace.example.name
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
