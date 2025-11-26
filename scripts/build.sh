#!/bin/bash

# 构建脚本 - 用于本地测试构建过程
# 用法: ./scripts/build.sh

set -e

echo "🚀 开始构建 webhookGo..."

# 创建输出目录
mkdir -p dist

# 构建配置
VERSION=${1:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}"

echo "📦 版本: ${VERSION}"
echo "⏰ 构建时间: ${BUILD_TIME}"
echo "🔧 Git提交: ${GIT_COMMIT}"
echo ""

# 支持的平台
platforms=(
    "linux/amd64"
    "linux/arm64"
    "linux/386"
    "linux/arm/7"
    "windows/amd64"
    "windows/386"
    "darwin/amd64"
    "darwin/arm64"
    "freebsd/amd64"
)

# 确保依赖正确
echo "📥 检查和更新依赖..."
go mod tidy
go mod verify

# 开始构建
for platform in "${platforms[@]}"; do
    IFS='/' read -r goos goarch goarm <<< "$platform"
    
    output_name="webhookGo-${goos}-${goarch}"
    if [ "$goos" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "🔨 构建 ${output_name}..."
    
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" GOARM="$goarm" \
    go build -ldflags="$LDFLAGS" -o "dist/$output_name" .
    
    # 创建压缩包
    cd dist
    if [ "$goos" = "windows" ]; then
        zip "${output_name%.exe}.zip" "$output_name"
        rm "$output_name"
    else
        tar -czf "${output_name}.tar.gz" "$output_name"
        rm "$output_name"
    fi
    cd ..
done

echo ""
echo "✅ 构建完成！"
echo "📁 输出目录: dist/"
echo "📋 文件列表:"
ls -la dist/