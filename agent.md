# Pipeline 治理規範與守則

> 本文件定義 GitHub Actions CI/CD Pipeline 的治理規則，適用於本專案所有 Workflow。

---

## 一、工具選用原則

### 1.1 僅限開源工具
- 所有 Action 與 CLI 工具必須為 **OSI 認可的開源授權**
- 允許：MIT、Apache 2.0、BSD-2-Clause、BSD-3-Clause、ISC、MPL-2.0
- 禁止：閉源工具、僅提供免費試用的商業工具、無明確授權的工具

### 1.2 無商業授權爭議
- 不採用授權條款模糊或定義不清的工具
- 不採用具有雙重授權爭議的工具（開源版功能受限，需付費解鎖）
- 不採用近期從開源變更為限制性授權的工具（BSL / SSPL / BUSL / ELv2）

### 1.3 供應鏈安全
- 所有第三方 Action 必須**鎖定版本標籤**（如 `@v4`）
- 禁止使用 `@master` 或 `@main` 等浮動分支
- 優先使用 GitHub 官方維護的 Action（`actions/*`）
- 定期審查第三方 Action 的授權與維護狀態

---

## 二、費用控管

| 項目 | 規則 |
|------|------|
| Runner | 僅使用 `ubuntu-latest`（GitHub-hosted） |
| Larger Runner | 禁止使用（需額外付費） |
| 第三方 SaaS | 禁止使用付費掃描服務 |
| Self-hosted Runner | 若使用需經評估並記錄成本 |

GitHub Actions 免費額度參考：

| 方案 | 每月免費分鐘數 | 儲存空間 |
|------|---------------|---------|
| Free | 2,000 分鐘 | 500 MB |
| Team | 3,000 分鐘 | 2 GB |
| Enterprise | 50,000 分鐘 | 50 GB |

---

## 三、資料安全

- 所有掃描與分析**皆在 Runner 本機執行**
- 原始碼、掃描結果、建置產物**不得上傳至 GitHub 以外的第三方服務**
- 建置產物僅透過 GitHub Artifact 暫存，僅限同 Workflow 內的 Job 存取

### 允許的外部通訊

| 工具 | 通訊對象 | 傳送內容 | 風險評估 |
|------|---------|---------|---------|
| govulncheck | vuln.go.dev（Go 官方） | module path | 低風險，不傳送原始碼 |
| setup-go | golang.org | Go SDK 版本號 | 低風險，僅下載 SDK |
| go mod download | proxy.golang.org | module path | 低風險，Go 官方 proxy |

### 資料流向

```
┌─────────────┐    checkout     ┌──────────────┐
│  GitHub Repo │ ─────────────→ │  Runner 本機  │
└─────────────┘                 │              │
                                │  build       │ → demo-app
                                │  test        │ → coverage.out
                                │  gosec       │ → 日誌輸出
                                │  govulncheck │ ─→ vuln.go.dev (僅 module path)
                                │  go-licenses │ → license-report.txt
                                └──────┬───────┘
                                       │ upload-artifact
                                       ▼
                                ┌──────────────┐
                                │ GitHub       │
                                │ Artifact     │ ← 僅限同 Workflow Job 存取
                                │ (暫存區)      │
                                └──────────────┘
```

---

## 四、第三方金鑰管理

- 目前 Pipeline **不需要任何第三方服務金鑰**
- 若未來新增需要金鑰的步驟，必須遵循：
  1. 在 `workflow_dispatch.inputs` 中明確宣告，讓使用者知悉
  2. 搭配 **GitHub Secrets** 使用，切勿硬編碼於 workflow 中
  3. 在本文件「工具授權清單」中記錄該服務的授權與資料政策

---

## 五、最小權限原則

Workflow 僅授予實際需要的 `GITHUB_TOKEN` 權限：

```yaml
permissions:
  contents: read
  actions: read
```

禁止使用 `permissions: write-all` 或省略 permissions。

---

## 六、工具授權清單

| 工具 | 版本 | 授權 | 維護者 | 資料外傳 |
|------|------|------|--------|---------|
| actions/checkout | @v5 | MIT | GitHub 官方 | 無 |
| actions/setup-go | @v5 | MIT | GitHub 官方 | 無 |
| actions/upload-artifact | @v4 | MIT | GitHub 官方 | 無（GitHub 內部） |
| actions/download-artifact | @v4 | MIT | GitHub 官方 | 無 |
| gosec (go install) | @latest | Apache 2.0 | securego 社群 | 無 |
| govulncheck | @latest | BSD-3-Clause | Go 官方 | module path → vuln.go.dev |
| go-licenses | @latest | Apache 2.0 | Google | 無 |

---

## 七、Pipeline 執行順序

```
build → deliver → test → security-scan → oss-scan
```

每個 Job 透過 `needs` 欄位串聯，確保嚴格按順序執行。

---

## 八、對話紀錄規範

- 所有 AI 對話紀錄存放於 `ai_logs/` 資料夾
- 檔名格式：`YYYYMMDD-HH_log.txt`（以小時為單位）
- 使用純文字格式，確保可讀性與通用性

### 自動記錄規則

LLM 在每次對話結束或階段性完成時，**必須自動執行以下步驟**：

1. **判斷當前時間**：取得當前日期與小時（格式 `YYYYMMDD-HH`）
2. **檢查是否已有對應 log 檔案**：
   - 若 `ai_logs/YYYYMMDD-HH_log.txt` 已存在 → **追加**新的對話紀錄至該檔案末尾
   - 若不存在 → **建立新檔案**
3. **跨小時處理**：若對話過程中跨越整點（如 09:58 開始、10:05 結束），則：
   - 在原小時的 log 檔案中記錄至該小時結束的內容
   - 在新小時的 log 檔案中記錄新小時開始後的內容
4. **紀錄格式**：
   ```
   HH:MM [角色] 內容摘要
   ```
   - 角色為 `使用者` 或 `AI`
   - 內容摘要應簡潔但完整，包含需求描述、執行動作、變更的檔案清單
5. **不可遺漏**：每一輪使用者提問與 AI 回應都必須記錄，不可跳過

---

## 變更紀錄

| 日期 | 變更內容 |
|------|---------|
| 2026-05-05 | 初版建立，定義工具選用、費用控管、資料安全、金鑰管理、最小權限、對話紀錄等規範 |
| 2026-05-05 | 對話紀錄規範新增自動記錄規則：LLM 須自動追加/建立 log、跨小時建立新檔、不可遺漏 |
