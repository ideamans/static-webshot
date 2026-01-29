# static-webshot - Static Web Screenshot Tool

Webページのピクセルベースのビジュアルリグレッションテストを行うCLIツールです。何らかの変更の前後でWebページのスクリーンショットを撮影し、ピクセル単位で比較することで表示上の見た目の変化を検出します。

実際のWebページにはカルーセルスライダーやCSSアニメーションなど、時間によって動的に変化する要素が多く含まれています。これらの要素があると、撮影のたびに異なるスクリーンショットが生成されるため、単純なピクセル比較では正確なリグレッションテストが困難です。static-webshotは、これらの非決定的な要因を自動的に抑制することで、動的な要素を多く含むページでも効率的かつ正確なピクセル比較によるビジュアルリグレッションテストを可能にするために開発されました。

## 特徴

- **決定論的スクリーンショット**: CSS/JSアニメーション、カルーセルスライダーの無効化、乱数の固定、時間の固定により、動的要素によるノイズを排除した一貫性のあるスクリーンショットを撮影
- **ピクセルベースのビジュアルリグレッション**: ベースラインと現在のスクリーンショットをピクセル単位で比較し、変化したピクセル数とパーセンテージを正確にレポート
- **デバイスプリセット**: デスクトップ・モバイル用のビューポート設定を内蔵
- **差分オーバーレイ出力**: 変化した領域をハイライトしたサイドバイサイドの差分画像を生成
- **柔軟なオプション**: ビューポート、リサイズ、マスキングなどをカスタマイズ可能

## インストール

```bash
go install github.com/ideamans/go-page-visual-regression-tester/cmd/staticwebshot@latest
```

またはソースからビルド:

```bash
git clone https://github.com/ideamans/go-page-visual-regression-tester.git
cd go-page-visual-regression-tester
go build -o static-webshot ./cmd/staticwebshot
```

## 使い方

### スクリーンショットの撮影

Webページのスクリーンショットを撮影:

```bash
# 基本的な使い方（デスクトッププリセット、1920x1080）
static-webshot capture https://example.com -o screenshot.png

# モバイルプリセット（390x844、iPhone User-Agent）
static-webshot capture https://example.com -o mobile.png --preset mobile

# カスタムビューポート
static-webshot capture https://example.com -o custom.png --viewport 1280x720

# 出力サイズをリサイズ
static-webshot capture https://example.com -o small.png --resize 800
static-webshot capture https://example.com -o thumb.png --resize 400x300

# ページ読み込み後に待機
static-webshot capture https://example.com -o loaded.png --wait-after 2000

# 特定の要素を非表示
static-webshot capture https://example.com -o clean.png --mask ".ad-banner" --mask ".cookie-notice"
```

### 画像の比較

2つのスクリーンショットを比較し、差分画像を生成:

```bash
# 基本的な比較
static-webshot compare baseline.png current.png -o diff.png

# テキスト形式でダイジェストを保存
static-webshot compare baseline.png current.png -o diff.png --digest-txt result.txt

# JSON形式でダイジェストを保存
static-webshot compare baseline.png current.png -o diff.png --digest-json result.json

# テキストとJSON両方
static-webshot compare baseline.png current.png -o diff.png --digest-txt result.txt --digest-json result.json
```

比較結果（差分パーセント含む）は常に標準出力に表示されます:

```
[Compare Result]
Baseline: baseline.png
Current: current.png
Output: ./diff.png
Diff Pixels: 100 / 100000
Diff Percent: 0.1000%
```

JSONダイジェスト出力 (`--digest-json`):

```json
{
  "pixelDiffCount": 100,
  "pixelDiffRatio": 0.001,
  "diffPercent": 0.1,
  "totalPixels": 100000,
  "baselinePath": "baseline.png",
  "currentPath": "current.png",
  "diffPath": "./diff.png"
}
```

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
| `--user-agent` | カスタムUser-Agent文字列（プリセットを上書き） | プリセット値 |
| `--headful` | ヘッドフルモードでブラウザを実行 | `false` |
| `--chrome-path` | Chrome実行ファイルのパス | 自動検出 |
| `-v, --verbose` | 詳細出力を有効化 | `false` |

## compareオプション

| オプション | 説明 | デフォルト |
|-----------|------|-----------|
| `-o, --output` | 差分画像の出力パス | `./diff.png` |
| `--digest-txt` | テキスト形式のダイジェスト出力パス | なし |
| `--digest-json` | JSON形式のダイジェスト出力パス | なし |
| `--color-threshold` | ピクセルごとの色差閾値（0-255） | `10` |
| `--ignore-antialiasing` | アンチエイリアスピクセルを無視 | `false` |
| `--label-font` | ラベル用TrueTypeフォントファイルのパス | 内蔵フォント |
| `--label-font-size` | ラベルのフォントサイズ（ポイント） | `14` |
| `--baseline-label` | baselineパネルのラベルテキスト | `baseline` |
| `--diff-label` | diffパネルのラベルテキスト | `diff` |
| `--current-label` | currentパネルのラベルテキスト | `current` |
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

つまり、Chromeがインストールされていない環境でもstatic-webshotを実行できます。その場合、Playwrightのブラウザ管理機能を使用してChromiumが自動的にダウンロードされます。

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
