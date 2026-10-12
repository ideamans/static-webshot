# 不具合: `--full-page` で、遅延読み込みの画像が白いまま写る（v0.4.0）

2026-10-11、ideamans-instructions のチュートリアル（リファクタリングの前後をページ全体で比べる）を撮っていて見つけた。**2026-10-11 修正（未リリース）**: 待ちを awaitPromise にし decode() も待つ・待ちに --timeout を掛ける・Loopback のログを抑止・Slick の無限ループで複製スライドが写る不具合も同時に修正。

## 症状

`loading="lazy"` の画像が、`--full-page` で撮ると **読み込まれる前の白い枠のまま写る**。何枚が白くなるかは撮るたびに変わる。

dummy-ec-site のソースのコピーを手元の PHP（Docker）で動かし、ファーストビューより下の部品の `<img>` に属性を足して、
元のままのページと `--full-page --preset mobile --mask .main-slider` で比べた結果（400×14760）:

| 足した属性 | 元のままとの差 |
|---|---|
| なし（元のまま同士を 2 回撮る） | 0.0000% |
| `decoding="async"` だけ | 0.0000% |
| `loading="lazy"` だけ | 1.4421% |
| `loading="lazy" decoding="async"` | 6.8708%（同じものをもう 1 回撮ると、1 回目との差が 8.3129%） |

白くなったのは MAGAZINE の段より下（ページの上から 5,845px 以降）の画像。

## 原因（確かめた）

`pkg/adapters/chromebrowser/browser.go` の `WaitForImages` と `WaitForFonts` は、Promise を返す式を `chromedp.Evaluate` で評価しているが、
**`chromedp.Evaluate` は既定では Promise の完了を待たない**（`runtime.EvaluateParams` の `awaitPromise` が false）。

chromedp v0.11.2（go.mod と同じ版）で `new Promise(r => setTimeout(r, 2000))` を評価して確かめた:

```text
Evaluate (static-webshot のまま): 2ms
Evaluate + WithAwaitPromise:      2.001s
```

つまり「画像を待つ」「フォントを待つ」は、実際には何も待っていない。

- 普通の画像は、`Navigate` が load イベントまで待つので、たまたま揃っている
- `--full-page` の `LoadLazyContent` が `loading="lazy"` を `eager` に替えた画像は、**その時点から読み込みが始まる**のに、待たずに撮る → 白い
- `decoding="async"` が付くと、読み込みのあとの描画も遅れるので、白い枚数が増える

## 直し方の見当

```go
chromedp.Evaluate(script, nil, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
	return p.WithAwaitPromise(true)
})
```

を `WaitForImages` と `WaitForFonts` の両方に入れる。あわせて:

- `WaitForImages` は `img.complete` のあとに `img.decode()` も待つと、`decoding="async"` の描画待ちまで含められる
  （`await Promise.all(imgs.map(i => i.decode().catch(() => {})))`）
- 待ちが本当に効くようになるので、壊れた画像や遅いサーバーで止まらないよう、`--timeout` の中で打ち切られることを確かめる
- 回帰テスト: `loading="lazy"` の画像を並べたページを `httptest` で配り、画像の応答をわざと遅らせて、`--full-page` で白くならないこと

## ついでに見つけたこと

**手元のサーバー（localhost）を撮ると、ログに `ERROR: could not unmarshal event: unknown IPAddressSpace value: Loopback` が百数十行出る。**
撮影は成功する。chromedp v0.11.2 / cdproto 2024-10 の版が、新しい Chrome の送ってくる値を知らないため。
chromedp と cdproto を上げれば消えるはず（未確認）。リファクタリングの前後確認は手元のサーバーで撮ることが多いので、
利用者が「失敗した」と思い込みやすい。

## 再現

```bash
~/dev/ideamans-instructions/labs/static-webshot/refactor/run.sh   # 18101 before / 18102 good / 18103 broken を立てて撮る
```

題材の作り方は同じフォルダの `prepare.py`。撮った画像は `labs/static-webshot/work-refactor/shots/`。
