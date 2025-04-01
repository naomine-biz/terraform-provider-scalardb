# Terraform Provider for ScalarDB

このTerraform Providerは、[ScalarDB](https://github.com/scalar-labs/scalardb)のDDL操作（テーブルの作成や変更など）をTerraformから管理するためのものです。

## 要件

- [Terraform](https://www.terraform.io/downloads.html) 1.0.x以上
- [Go](https://golang.org/doc/install) 1.18以上（開発時のみ）
- [ScalarDB](https://github.com/scalar-labs/scalardb) サーバー

## インストール

### Terraform 0.13以上の場合

```hcl
terraform {
  required_providers {
    scalardb = {
      source  = "scalar-labs/scalardb"
      version = "0.1.0"
    }
  }
}

provider "scalardb" {
  # 設定オプション
}
```

### 手動インストール

1. [リリースページ](https://github.com/scalar-labs/terraform-provider-scalardb/releases)から最新のバイナリをダウンロードします。
2. プラグインディレクトリに解凍します：
   ```
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/scalar-labs/scalardb/0.1.0/[OS]_[ARCH]/
   mv terraform-provider-scalardb_v0.1.0 ~/.terraform.d/plugins/registry.terraform.io/scalar-labs/scalardb/0.1.0/[OS]_[ARCH]/
   ```
3. バイナリに実行権限を付与します：
   ```
   chmod +x ~/.terraform.d/plugins/registry.terraform.io/scalar-labs/scalardb/0.1.0/[OS]_[ARCH]/terraform-provider-scalardb_v0.1.0
   ```

## 使用方法

```hcl
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
```

## プロバイダーの設定

| 名前 | 説明 | タイプ | デフォルト | 必須 |
|------|-------------|------|---------|:--------:|
| host | ScalarDBサーバーのホストアドレス | `string` | n/a | はい |
| port | ScalarDBサーバーのポート | `number` | `60051` | いいえ |
| username | ScalarDB認証用のユーザー名 | `string` | n/a | いいえ |
| password | ScalarDB認証用のパスワード | `string` | n/a | いいえ |

## リソース

### scalardb_namespace

名前空間（Namespace）を管理します。

#### 引数

| 名前 | 説明 | タイプ | デフォルト | 必須 |
|------|-------------|------|---------|:--------:|
| name | 名前空間の名前 | `string` | n/a | はい |
| replication_factor | レプリケーション係数 | `number` | `1` | いいえ |
| strategy_class | レプリケーション戦略クラス | `string` | `"SimpleStrategy"` | いいえ |
| durable_writes | 永続的な書き込みを使用するかどうか | `bool` | `true` | いいえ |

### scalardb_table

テーブルを管理します。

#### 引数

| 名前 | 説明 | タイプ | デフォルト | 必須 |
|------|-------------|------|---------|:--------:|
| namespace | テーブルを作成する名前空間 | `string` | n/a | はい |
| name | テーブルの名前 | `string` | n/a | はい |
| partition_key | テーブルのパーティションキー列 | `list(string)` | n/a | はい |
| clustering_key | テーブルのクラスタリングキー列 | `list(string)` | `[]` | いいえ |
| column | テーブルの列定義 | `set(object)` | n/a | はい |
| compaction_strategy | コンパクション戦略 | `string` | `"SizeTieredCompactionStrategy"` | いいえ |
| clustering_order | クラスタリング順序 | `map(string)` | `{}` | いいえ |

#### column引数

| 名前 | 説明 | タイプ | デフォルト | 必須 |
|------|-------------|------|---------|:--------:|
| name | 列の名前 | `string` | n/a | はい |
| type | 列のデータ型（INT, BIGINT, TEXT, FLOAT, DOUBLE, BOOLEAN, BLOB） | `string` | n/a | はい |

## 開発

### 必要条件

- [Terraform](https://www.terraform.io/downloads.html) 1.0.x以上
- [Go](https://golang.org/doc/install) 1.18以上

### ビルド

1. リポジトリをクローンします：
   ```
   git clone https://github.com/scalar-labs/terraform-provider-scalardb.git
   ```

2. プロバイダーをビルドします：
   ```
   go build
   ```

### テスト

#### ユニットテスト

```
go test ./...
```

#### 統合テスト

モックのScalarDB Clusterサーバーに対してTerraformのplan/applyを実行するテストを実行できます：

```
cd tests
./run_test.sh
```

テストスクリプトは以下のオプションをサポートしています：

- `--port <port>`: モックサーバーのポート番号（デフォルト: 60051）
- `--provider-path <path>`: プロバイダーのパス（デフォルト: ../terraform-provider-scalardb）
- `--log-level <level>`: Terraformのログレベル（デフォルト: INFO）

例：
```
./run_test.sh --port 60052 --log-level DEBUG
```

ScalarDB Clusterのエンドポイントは環境変数で設定することもできます：

```
export TF_VAR_scalardb_host=localhost
export TF_VAR_scalardb_port=60051
export TF_VAR_scalardb_username=admin
export TF_VAR_scalardb_password=password
```

### インストール（開発用）

```
go build
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/scalar-labs/scalardb/0.1.0/[OS]_[ARCH]/
cp terraform-provider-scalardb ~/.terraform.d/plugins/registry.terraform.io/scalar-labs/scalardb/0.1.0/[OS]_[ARCH]/
```

## ライセンス

[Apache License 2.0](LICENSE)
