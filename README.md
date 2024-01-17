go-htnblog：はてなブログ-クライアント
====================================

[![GoDev](https://pkg.go.dev/badge/github.com/hymkor/go-htnblog)](https://pkg.go.dev/github.com/hymkor/go-htnblog)
[![Github latest Releases](https://img.shields.io/github/downloads/hymkor/go-htnblog/latest/total.svg)](https://github.com/hymkor/go-htnblog/releases/latest)

本パッケージは

- コマンドラインクライアント(htnblog.exe, htnblog)  
    → シェル、コマンドプロンプトなどから利用。vim など任意のテキストエディターで記事を編集して、投稿
- クライアントライブラリ(github.com/hymkor/go-htnblog)  
    → Go言語で投稿機能を利用できるライブラリ

からなります。Windows, Linux (Ubuntu on WSL) で動作を確認しています


コマンドラインクライアント
--------------------------

### インストール

[Releases](https://github.com/hymkor/go-htnblog/releases)
よりダウンロードして、実行ファイルを展開してください

#### scoop インストーラーがある場合

```
scoop install https://raw.githubusercontent.com/hymkor/go-htnblog/master/htnblog.json
```

もしくは

```
scoop bucket add hymkor https://github.com/hymkor/scoop-bucket
scoop install htnblog
```

### 設定

`htnblog init` を実行してください。

```
$ htnblog init
Hatena-id ? (← Hatena-IDを入力)
API-KEY ? (← APIキー入力)
End Point URL 1 ? (← 一つ目のブログのEndPointURL を入力)
End Point URL 2 ? (← 二つ目のブログのEndPointURL を入力:Enterのみで省略可)
End Point URL 3 ? (← 三つめのブログのEndPointURL を入力:Enterのみで省略可)
End Point URL (default) ? (← デフォルトのブログの EndPointURL: 既定値は1と同じ)
Text Editor Path ? (← テキストエディタのパス: 既定値は%EDITOR%)
Saved configuration to C:\Users\hymkor\.htnblog
```

設定は ~/.htnblog というファイルに次のように保存されます。

```
{
    "userid":"(YOUR_HATENA_ID)",
    "endpointurl":"(END_POINT_URL)",
    "apikey":"(YOUR API KEY)",
    "editor":"(YOUR EDITOR FULLPATH))"
    "endpointurl1":"(END_POINT_URL at option `-1`)",
    "endpointurl2":"(END_POINT_URL at option `-2`)",
    "endpointurl3":"(END_POINT_URL at option `-3`)",
}
```

- **endpointurl** は 「はてな」の「ダッシュボード」 → (ブログ名) → 「設定」 → 詳細設定 → 「AtomPub」セクションのルートネンドポイントの記載の URL になります。
- **apikey** は「アカウント設定」→「APIキー」に記載されています。
- **editor** は編集に使うテキストエディターのパスを記載してください。
- **endpointurl[123]** は **endpointurl** と同じですが、オプション -1,-2,-3 が指定された時にこちらが使われます。

### 使い方

[cmd/htnblog/main.go](cmd/htnblog/main.go)

- `htnblog` (オプションなし) … ヘルプ
- `htnblog init` … 設定の編集
- `htnblog list` … 直近10件の記事のリスト
- `htnblog new` … 新規記事のドラフトを作成
- `htnblog edit {URL|@0|@1|…}` … 既存記事の編集
- `htnblog publish {URL|@0|@1|…}` … 指定した記事を下書き状態から公開状態へ変更する

```./htnblog |
htnblog v0.7.0-13-gea7ce11-windows-amd64 by go1.21.5

Usage: htnblog {options...} {init|list|new|type|edit}
  htnblog init                   ... edit configuration
  htnblog list                   ... show recent articles
  htnblog new                    ... create a new draft
  htnblog type    {URL|@0|@1|..} ... output the article to STDOUT
  htnblog edit    {URL|@0|@1|..} ... edit the article
  htnblog delete  {URL|@0|@1|..} ... output the article to STDOUT and delete it
  htnblog publish {URL|@0|@1|..} ... set false the draft flag of the article
    The lines in the draft up to "---" are the header lines,
    and the rest is the article body.

  -1	Use the value of "endpointurl1" in the JSON setting
  -2	Use the value of "endpointurl2" in the JSON setting
  -3	Use the value of "endpointurl3" in the JSON setting
  -f	Delete without prompt
  -n int
    	fetch articles (default 100)
  -rc string
    	use the specified file instead of ~/.htnblog
```

Goライブラリ go-htnblog
------------------------

### 使用例: 一覧表示

[examples/list.go](examples/list.go)

設定は標準入力から読み込むようにしてます。ブログに投稿するためのアカウント情報を保持する htnblog.Blog型のインスタンスは `htnblog.NewFromJSON` 関数で JSON テキストから生成していますが、別に `&Blog{...}` でいきなり生成しても構いません。

```examples/list.go
package main

import (
    "fmt"
    "io"
    "os"

    "github.com/hymkor/go-htnblog"
)

func list() error {
    auth, err := io.ReadAll(os.Stdin)
    if err != nil {
        return err
    }
    blog, err := htnblog.NewFromJSON(auth)
    if err != nil {
        return err
    }
    entries, err := blog.List()
    if err != nil {
        return err
    }
    for _, entry1 := range entries {
        fmt.Println(entry1.Title)
        fmt.Println(entry1.EditUrl())
    }
    return nil
}

func main() {
    if err := list(); err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        os.Exit(1)
    }
}
```

### 使用例: 新規投稿

[examples/post.go](examples/post.go)

`(*Blog) Post` 関数は戻り値としてウェブアクセスの返信を io.ReadCloser と error で返してきます。 いちいち処理するのが面倒なので `htnblog.Dump` という関数を用意して、そのまま標準出力に転送して Close させています。

```examples/post.go
package main

import (
    "fmt"
    "io"
    "os"
    "time"

    "github.com/hymkor/go-htnblog"
)

func post() error {
    auth, err := io.ReadAll(os.Stdin)
    if err != nil {
        return err
    }
    blog, err := htnblog.NewFromJSON(auth)
    if err != nil {
        return err
    }
    return htnblog.Dump(blog.Post(time.Now().Format("投稿 2006-01-02 15:04:05"), "本文を書く", "yes"))
}

func main() {
    if err := post(); err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        os.Exit(1)
    }
}
```

### 使用例: 最も新しい記事を編集

[examples/edit.go](examples/edit.go)

```examples/edit.go
package main

import (
    "errors"
    "fmt"
    "io"
    "os"
    "time"

    "github.com/hymkor/go-htnblog"
)

func edit() error {
    auth, err := io.ReadAll(os.Stdin)
    if err != nil {
        return err
    }
    blog, err := htnblog.NewFromJSON(auth)
    if err != nil {
        return err
    }
    entries, err := blog.List()
    if err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        os.Exit(1)
    }
    if len(entries) <= 0 {
        return errors.New("no entries")
    }
    entries[0].Content.Body += time.Now().Format("\n編集 2006-01-02 15:04:05")
    return htnblog.Dump(blog.Update(entries[0]))
}

func main() {
    if err := edit(); err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        os.Exit(1)
    }
}
```

Refrences
---------

* [はてなブログAtomPub - はてなブログ ヘルプ](https://help.hatenablog.com/entry/atompub)
* [はてなブログAtomPub | Hatena Developer Center](https://developer.hatena.ne.jp/ja/documents/blog/apis/atom/)
