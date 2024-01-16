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
