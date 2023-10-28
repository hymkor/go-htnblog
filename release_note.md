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
