# EasySub2api 手动发布

EasySub2api 不会因分支 push 或 tag push 自动发布。分支 push/PR 仍运行 CI 和安全扫描，但只有维护者在本地执行发布命令后，GitHub 才会构建 GHCR 镜像并创建 Release。

## 发布产物

- `ghcr.io/asukacc/easysub2api:X.Y.Z`
- 同步更新 `X.Y`、`X` 和 `latest` 标签。
- 单一 manifest 同时支持 `linux/amd64` 与 `linux/arm64`。
- GitHub Release 只包含版本说明，不包含二进制附件。
- 不发布 Docker Hub 镜像。

## 前置条件

```powershell
winget install GitHub.cli
gh auth login
gh auth status
```

还需要 Git、PowerShell 7，并确保当前仓库处于干净、已同步的 `main` 分支。

## 发布命令

自动从上一个 tag 之后的提交生成说明：

```powershell
pwsh ./scripts/release.ps1 -Version 0.0.2
```

使用人工准备的说明：

```powershell
pwsh ./scripts/release.ps1 -Version 0.0.2 -NotesFile ./release-notes.md
```

脚本会更新 VERSION、提交并推送 `main`，等待 `CI` 与 `Security Scan` 成功，创建 annotated tag，随后触发并等待 `release.yml`。版本只允许稳定的 `X.Y.Z`。

## 失败与重试

- CI 或安全扫描失败：脚本停止且不会创建 tag；修复提交后使用新的版本号重新发布。
- tag 已存在但指向其他提交：脚本停止，不覆盖或删除 tag。
- 发布工作流失败：修复工作流或构建环境后，用相同版本再次执行命令；脚本会复用一致的版本提交和 tag。
- GitHub Release 已存在：脚本视为已完成，不重复发布。

不要直接在 GitHub 网页中创建版本 tag。`workflow_dispatch` 虽然可见，但正常发布必须通过本地脚本，以确保 VERSION、检查结果和 tag 一致。
