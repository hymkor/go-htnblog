- htnblog.exe:
    - `htnblog new` で Category: ヘッダーが認識されない問題を修正
    - `-debug` オプションが機能しなくなっている問題を修正
- go-htnblog:
    - 新規エントリ追加用で、Category フィールドなども指定ができる `(*Blog) Add` を追加

v1.0.0
======
(2024.02.08)

- htnblog.exe:
    - `Category: PowerShell Windows11` といった体裁でカテゴリーを編集できるようにした。
    - `new`,`edit`: 終了前に編集用URL、表示URLを表示するようにした
- go-htnblog:
    - `(*XmlEntry) UrlForBrowserToEdit` を追加
    - `(*XmlEntry) EditUrl` を `UrlToEdit` に改名 (旧名は残しているが Deprecated )

v0.9.0
======
(2024.01.20)

- htnblog.exe:
    - `htnblog init` 変更前の .htnblog を .htnblog~ にバックアップしておくようにした
    - `htnblog init` (linux) 保存する .htnblog のパーミッションを 0644 ではなく、0600 とした
    - 編集ページをウェブブラウザで開くサブコマンド: `htnblog browse` を追加
    - `htnblog list` で `-n` オプションで指定されなかった場合の表示記事数を 100 から 10 に変更した

v0.8.0
======
(2024.01.18)

- htnblog.exe:
    - オプション `-updated` を削除。今後、記事の日付を変更する場合は、エディター上で、ヘッダ行 `Updated:` の日付を変更する
    - サブコマンド: `publish` を追加。指定した記事を下書き状態から公開状態へ変更する
    - 処理実行後、サーバーレスポンスを全て表示していたのを、オプション `-debug` が設定されない限りはステイタスコード出力だけにするようにした
    - サブコマンド: `init`: %EDITOR% が未設定の場合、候補として、notepad.exe や /etc/alternatives/editor を入力時の既定値とするようにした。
- **ライブラリ互換性の破壊的変更**: `(*Blog) Post`, `(*Blog) Update`, `(*Blog) Delete` などのパラメータ・戻り値を変更 (※ 戻り値を `io.ReadCloser` から `*http.Response` へ変更するなど)

v0.7.0
======
(2024.01.17)

- `htnblog edit` で変更前項目の値をコメントの形でドラフトに組み込みようにした
- 記事の日付を変えられるように、Updated: ヘッダをドラフトに用意
- 変更できない記事データも、参考のためドラフトの Rem: ヘッダに列挙するようにした
- `htnblog init` で設定の編集をできるようにした

v0.6.0
======
(2023.12.27)

- htnblog.exe: `-1`,`-2`,`-3` で別のルートエンドポイントURLを使えるようにした
  (それぞれ設定ファイルの `"endpointurl1"`, `"endpointurl2"`, `"endpointurl3"` を参照)
- htnblog.exe: 記事を削除する `htnblog delete` サブコマンドと`-f`オプションを追加
- メソッド: `(*Blog) Delete` を実装

v0.5.1
======
(2023.12.26)

- エントリIDのない `htnblog edit` は最後のエントリを編集すべきだったのに、エラー終了していた動作を修正

v0.5.0
======
(2023.12.23)

- `htnblog list` で10個以上のエントリを表示可能とした
    - `-n` で変更可能。デフォルトは `-n 100`
- `htnblog edit @N` で N を 10以上にできるようにした
- `htnblog edit (記事のURL)` をサポート
- `(*Blog) Index`, `(*Blog) EachEntry`, `(*XmlEntry) AlternateUrl` を実装

v0.4.0
======
(2023.11.11)

- 一時ファイル名にプロセスIDを使うようにした (`$TEMP/htnblog-(PID).md`)
- 投稿前に `Are you sure to post ? (Yes/No/Edit):` と尋ねるようにした
- 使用済みのドラフトテキストファイルは、即削除ではなく、OSのゴミ箱に移動するようにした  
  (非Windowsでは、 [freedesktop.org のデスクトップのゴミ箱仕様](https://www.freedesktop.org/wiki/Specifications/trash-spec/) に準拠した形で "the home trash" へ移動)
- 投稿済み記事を標準出力に出す `htnblog.exe type [article-no]` を実装

v0.3.1
======
(2023.10.28)

- ドラフトのテキストが空だった時、投稿をキャンセルするようにした
- `htnblog.exe edit` で、ドラフトフラグが no になってしまう不具合を修正

v0.3.0
======
(2023.10.18)

- `htnblog.exe list`: URLではなく entry-id を表示するようにした
- `htnblog.exe`: 引数が与えられなかった時だけバージョンを表示するようにした
- `htnblog list`, `htnblog edit` で @0～@9 を entry-id の別名として使えるようにした
- `htnblog.exe`: .htnblog のかわりに使うファイルを指定する -rc FILENAME オプションを追加
- 設定からフィールド author を削除。userid を使用するようにした

v0.2.0
======
(2023.10.17)

- `htnblog edit`:
    - 編集した記事の日付が今日になってしまう問題を修正
    - 編集記事のエントリIDを指定できるようにした
    - 日付がおかしくなった記事を修正するため `htnblog -updated yyyy-mm-ddThh:mm:ss+09:00 edit entry-id` のように日付を変更できるようにした

v0.1.1
======
(2023.10.16)

- 公開済みの記事に対して `htnblog edit` を実行すると、投稿時に
  `400 Cannot Change into Draft` というエラーになってしまう不具合を修正
- 起動時にバージョンを表示するようにした

v0.1.0
------
(2023.10.16)

- 初版
