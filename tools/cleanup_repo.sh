#!/usr/bin/env bash
set -euo pipefail

# --- 可调参数 ---
BRANCH=${1:-main}            # 目标分支（默认 main）
COMMIT_MSG="chore: cleanup junk & update .gitignore"

# --- 要从 Git 索引里移除（若已追踪）的路径/模式 ---
TRACKED_PATTERNS=(
  "webapi"
  "cmd/webapi/webapi"
  "bin"
  "build"
  "dist-webapi"
  "webui/dist"
  "node_modules"
  "**/node_modules"
  "**/package-lock.json"
  "SELECT"
  "Using"
  "test.png"
  "restart_webapi.sh"
  "open-node.sh"
  "*.db"
  "*.db.bak.*"
  "vet.log"
)

# --- 仅本地删除的垃圾（未必被追踪） ---
LOCAL_ONLY=(
  "SELECT"
  "Using"
  "test.png"
  "restart_webapi.sh"
  "open-node.sh"
  "vet.log"
  "*.db"
  "*.db.bak.*"
)

echo "👉 切到分支: ${BRANCH}"
git checkout "${BRANCH}" >/dev/null 2>&1 || true

echo "👉 确保 .gitignore 已包含最新规则（如尚未更新，请先覆盖为我们整理的版本）"

# 1) 从 Git 索引移除已追踪的构建产物/垃圾
echo "🧹 从 Git 索引移除已追踪文件…(忽略未匹配项的报错)"
for pat in "${TRACKED_PATTERNS[@]}"; do
  # 用 git ls-files 判断是否被追踪；若有则 git rm --cached
  if git ls-files -z -- "$pat" | grep -q .; then
    echo "  - git rm --cached: $pat"
    # -r 递归, -f 强制（有时二进制/改动需要）
    git rm --cached -r -f -- "$pat" || true
  fi
done

# 2) 添加 .gitignore 和其它改动
echo "➕ git add 变更与 .gitignore"
git add .gitignore || true
git add -A || true

# 3) 提交（若有改动）
if ! git diff --cached --quiet; then
  echo "✅ 提交：${COMMIT_MSG}"
  git commit -m "${COMMIT_MSG}"
else
  echo "ℹ️ 没有需要提交的索引改动。"
fi

# 4) 推送
echo "🚀 推送到远程 ${BRANCH}"
git push origin "${BRANCH}" || true

# 5) 本地物理删除垃圾文件/目录
echo "🧽 删除本地垃圾文件/目录…"
# 先删固定路径
rm -rf webapi cmd/webapi/webapi bin build dist-webapi webui/dist node_modules 2>/dev/null || true
# 再删通配
for pat in "${LOCAL_ONLY[@]}"; do
  find . -type f -name "${pat}" -print -delete 2>/dev/null || true
done
# delete package-lock.json（任意子目录）
find . -type f -name "package-lock.json" -print -delete 2>/dev/null || true

echo "🎉 清理完成！"
