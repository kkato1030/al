#!/bin/bash

set -e

# アーキテクチャを自動検出
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
  ARCH="arm64"
else
  echo "Error: Unsupported architecture: $ARCH"
  exit 1
fi

# バージョン指定（環境変数で上書き可能）
VERSION="${AL_VERSION:-latest}"

# 最新バージョンを取得
if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -s https://api.github.com/repos/kkato1030/al/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

# インストール先のディレクトリ
INSTALL_DIR="${AL_INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="al"

echo "Installing al ${VERSION} for darwin/${ARCH}..."

# 一時ディレクトリを作成
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# ダウンロード
DOWNLOAD_URL="https://github.com/kkato1030/al/releases/download/${VERSION}/al_darwin_${ARCH}.tar.gz"
echo "Downloading from ${DOWNLOAD_URL}..."
curl -L -o "${TMP_DIR}/al.tar.gz" "${DOWNLOAD_URL}"

# 解凍
echo "Extracting..."
tar -xzf "${TMP_DIR}/al.tar.gz" -C "${TMP_DIR}"

# インストール先ディレクトリが存在しない場合は作成
if [ ! -d "$INSTALL_DIR" ]; then
  echo "Creating directory: $INSTALL_DIR"
  sudo mkdir -p "$INSTALL_DIR"
fi

# バイナリをインストール
echo "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
sudo mv "${TMP_DIR}/al" "${INSTALL_DIR}/${BINARY_NAME}"
sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# インストール確認
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
  echo "Installation completed successfully!"
  echo ""
  "$BINARY_NAME" version
else
  echo "Warning: Installation may have failed. Please check if ${INSTALL_DIR} is in your PATH."
  exit 1
fi
