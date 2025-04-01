#!/bin/bash

# エラー時に停止し、未定義の変数を使用した場合にエラーを表示
set -e
set -u

# デバッグモードを有効にする（すべてのコマンドを表示）
set -x

# デフォルト値
MOCK_PORT=60051
PROVIDER_PATH="../terraform-provider-scalardb"
TERRAFORM_LOG=INFO

# 引数の解析
while [[ $# -gt 0 ]]; do
    case $1 in
    --port)
        MOCK_PORT="$2"
        shift 2
        ;;
    --provider-path)
        PROVIDER_PATH="$2"
        shift 2
        ;;
    --log-level)
        TERRAFORM_LOG="$2"
        shift 2
        ;;
    *)
        echo "Unknown option: $1"
        exit 1
        ;;
    esac
done

# 現在のディレクトリを保存
CURRENT_DIR=$(pwd)

# スクリプトのディレクトリに移動
cd "$(dirname "$0")"
echo "Current directory: $(pwd)"

# モックサーバーをビルド
echo "Building mock server..."
go build -o mock_server ./mock/mock_server.go || {
    echo "Failed to build mock server"
    exit 1
}
echo "Mock server built successfully"

# プロバイダーをビルド
echo "Building provider..."
cd ..
echo "Current directory: $(pwd)"
go build -o terraform-provider-scalardb || {
    echo "Failed to build provider"
    exit 1
}
echo "Provider built successfully"

# プロバイダーの絶対パスを取得
PROVIDER_ABS_PATH=$(pwd)/terraform-provider-scalardb
echo "Provider path: $PROVIDER_ABS_PATH"

cd tests
echo "Current directory: $(pwd)"

# Terraformプラグインディレクトリを作成
PLUGIN_DIR="$HOME/.terraform.d/plugins/registry.terraform.io/scalar-labs/scalardb/0.1.0/$(go env GOOS)_$(go env GOARCH)"
echo "Plugin directory: $PLUGIN_DIR"
mkdir -p "$PLUGIN_DIR"
cp "$PROVIDER_ABS_PATH" "$PLUGIN_DIR/" || {
    echo "Failed to copy provider to plugin directory"
    exit 1
}
echo "Provider copied to plugin directory"

# モックサーバーを起動
echo "Starting mock server on port $MOCK_PORT..."
export SCALARDB_MOCK_PORT=$MOCK_PORT
./mock_server &
MOCK_PID=$!
echo "Mock server started with PID: $MOCK_PID"

# モックサーバーが終了したらクリーンアップ
function cleanup {
    echo "Stopping mock server..."
    kill $MOCK_PID || true
    echo "Mock server stopped"
}
trap cleanup EXIT

# モックサーバーが起動するのを待つ
echo "Waiting for mock server to start..."
sleep 5
echo "Mock server should be ready now"

# Terraformの環境変数を設定
export TF_VAR_scalardb_host=localhost
export TF_VAR_scalardb_port=$MOCK_PORT
export TF_VAR_scalardb_username=admin
export TF_VAR_scalardb_password=password
export TF_LOG=$TERRAFORM_LOG
echo "Terraform environment variables set"

# Terraformを初期化
echo "Initializing Terraform..."
terraform init -upgrade || {
    echo "Failed to initialize Terraform"
    exit 1
}
echo "Terraform initialized successfully"

# Terraformプラン
echo "Running Terraform plan..."
terraform plan -out=tfplan || {
    echo "Failed to create Terraform plan"
    exit 1
}
echo "Terraform plan created successfully"

# Terraformを適用
echo "Applying Terraform plan..."
terraform apply -auto-approve tfplan || {
    echo "Failed to apply Terraform plan"
    exit 1
}
echo "Terraform plan applied successfully"

# Terraformを破棄
echo "Destroying Terraform resources..."
terraform destroy -auto-approve || {
    echo "Failed to destroy Terraform resources"
    exit 1
}
echo "Terraform resources destroyed successfully"

echo "Test completed successfully!"

# 元のディレクトリに戻る
cd "$CURRENT_DIR"
echo "Returned to original directory: $(pwd)"
