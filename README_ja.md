# static-web-shot - Static Web Screenshot Tool

決定論的なスクリーンショットを撮影し、視覚的回帰テストのために比較するCLIツールです。

## 特徴

- **決定論的スクリーンショット**: アニメーションの無効化、乱数の固定、時間の固定により、一貫したスクリーンショットを撮影
- **デバイスプリセット**: デスクトップ・モバイル用のビューポート設定を内蔵
- **視覚比較**: ピクセル単位の画像比較と差分オーバーレイ出力
- **柔軟なオプション**: ビューポート、リサイズ、マスキングなどをカスタマイズ可能

## インストール

```bash
go install github.com/ideamans/go-page-visual-regression-tester/cmd/static-web-shot@latest
```

またはソースからビルド:

```bash
git clone https://github.com/ideamans/go-page-visual-regression-tester.git
cd go-page-visual-regression-tester
go build -o static-web-shot ./cmd/static-web-shot
```

## 使い方

### スクリーンショットの撮影

Webページのスクリーンショットを撮影:

```bash
# 基本的な使い方（デスクトッププリセット、1920x1080）
static-web-shot capture https://example.com -o screenshot.png

# モバイルプリセット（390x844、iPhone User-Agent）
static-web-shot capture https://example.com -o mobile.png --preset mobile

# カスタムビューポート
static-web-shot capture https://example.com -o custom.png --viewport 1280x720

# 出力サイズをリサイズ
static-web-shot capture https://example.com -o small.png --resize 800
static-web-shot capture https://example.com -o thumb.png --resize 400x300

# ページ読み込み後に待機
static-web-shot capture https://example.com -o loaded.png --wait-after 2000

# 特定の要素を非表示
static-web-shot capture https://example.com -o clean.png --mask ".ad-banner" --mask ".cookie-notice"
```

### 画像の比較

2つのスクリーンショットを比較し、差分画像を生成:

```bash
# 基本的な比較
static-web-shot compare baseline.png current.png -o diff.png

# カスタム閾値（0.0-1.0）
static-web-shot compare baseline.png current.png -o diff.png --threshold 0.1
```

終了コード:
- `0`: 閾値内で一致（PASS）
- `1`: 閾値を超えて差異あり（FAIL）

## captureオプション

| オプション | 説明 | デフォルト |
|-----------|------|-----------|
| `-o, --output` | 出力ファイルパス | `./capture.png` |
| `--preset` | デバイスプリセット（`desktop`, `mobile`） | `desktop` |
| `--viewport` | ビューポートサイズ（`幅x高さ` または `幅`） | プリセット値 |
| `--resize` | 出力画像サイズ（`幅x高さ` または `幅`） | リサイズなし |
| `--wait-after` | ページ読み込み後の待機時間（ms） | `0` |
| `--mask` | 非表示にする要素のCSSセレクタ（複数指定可） | なし |
| `--wait-selector` | 待機するCSSセレクタ（複数指定可） | なし |
| `--inject-css` | 注入するカスタムCSS | なし |
| `--mock-time` | Date APIの固定時刻（ISO 8601形式） | なし |
| `--proxy` | HTTPプロキシURL | なし |
| `--ignore-tls-errors` | TLS証明書エラーを無視 | `false` |
| `--timeout` | ナビゲーションタイムアウト（秒） | `30` |
| `--headful` | ヘッドフルモードでブラウザを実行 | `false` |
| `--chrome-path` | Chrome実行ファイルのパス | 自動検出 |
| `-v, --verbose` | 詳細出力を有効化 | `false` |

## compareオプション

| オプション | 説明 | デフォルト |
|-----------|------|-----------|
| `-o, --output` | 差分画像の出力パス | `./diff.png` |
| `--threshold` | 許容するピクセル差異率（0.0-1.0） | `0.15` |
| `--color-threshold` | ピクセルごとの色差閾値（0-255） | `10` |
| `--ignore-antialiasing` | アンチエイリアスピクセルを無視 | `false` |
| `-v, --verbose` | 詳細出力を有効化 | `false` |

## デバイスプリセット

| プリセット | ビューポート | User-Agent |
|-----------|-------------|------------|
| `desktop` | 1920x1080 | Windows Chrome |
| `mobile` | 390x844 | iPhone Safari |

## Chromeの自動検出

ツールは以下の優先順位でChromeを自動的に検出します:
1. `--chrome-path` オプション（明示的なパス指定）
2. `CHROME_PATH` 環境変数
3. システムにインストールされたChrome/Chromium
4. **Playwright経由で自動インストール**（Chromeが見つからない場合）

つまり、Chromeがインストールされていない環境でもstatic-web-shotを実行できます。その場合、Playwrightのブラウザ管理機能を使用してChromiumが自動的にダウンロードされます。

## 決定論的な処理

一貫したスクリーンショットを確保するため、以下の処理が自動的に適用されます:

**JavaScript修正:**
- `Date.now()` と `new Date()` を固定値に
- `Math.random()` を常に0.5を返すように固定
- `Performance.now()` を固定値に
- 動画・音声の自動再生を無効化
- IntersectionObserverで全要素を可視状態に（遅延読み込み対策）
- スクロール関連の動作を無効化
- Web Animations APIを無効化

**CSS修正:**
- 全てのCSSアニメーション・トランジションを無効化
- テキストカーソル（キャレット）を非表示
- スムーズスクロールを無効化

## ライセンス

MIT License
