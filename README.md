go-htnblog：はてなブログ投稿用Go言語パッケージ
==============================================

[![GoDev](https://pkg.go.dev/badge/github.com/hymkor/go-htnblog)](https://pkg.go.dev/github.com/hymkor/go-htnblog)

まだ BASIC 認証でごめん

サンプル投稿ツール (htnblog)
----------------------------

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

### 設定ファイル用意

~/.htnblog というファイルに次のように記載してください

```
{
    "userid":"(YOUR_HATENA_ID)",
    "endpointurl":"(END_POINT_URL)",
    "apikey":"(YOUR API KEY)",
    "editor":"(YOUR EDITOR FULLPATH))"
}
```

- **endpointurl** は 「はてな」の「ダッシュボード」 → (ブログ名) → 「設定」 → 詳細設定 → 「AtomPub」セクションのルートネンドポイントの記載の URL になります。
- **apikey** は「アカウント設定」→「APIキー」に記載されています。
- **editor** は編集に使うテキストエディターのパスを記載してください。

### 使い方

[cmd/htnblog/main.go](cmd/htnblog/main.go)

- `htnblog` (オプションなし) … ヘルプ
- `htnblog list` … 直近10件の記事のリスト
- `htnblog new` … 新規記事のドラフトを作成
- `htnblog edit {URL|@0|@1|…}` … 既存記事の編集

```./htnblog |
htnblog v0.5.0-windows-amd64 by go1.21.5

Usage: htnblog {list|new|type|edit}
  htnblog list                ... show recent articles
  htnblog new                 ... create a new draft
  htnblog type {URL|@0|@1|..} ... output the article to STDOUT
  htnblog edit {URL|@0|@1|..} ... edit the article
    The lines in the draft up to "---" are the header lines,
    and the rest is the article body.

Please write your setting on ~/.htnblog as below:
    {
        "userid":"(YOUR_USER_ID)",
        "endpointurl":"(END_POINT_URL)",
        "apikey":"(YOUR API KEY)",
        "editor":"(YOUR EDITOR.THIS IS for cmd/htnblog/main.go)"
    }

  -n int
    	fetch articles (default 100)
  -rc string
    	use the specified file instead of ~/.htnblog
  -updated string
    	(experimental) set the updated date like 2006-01-02T15:04:05-07:00
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
    return htnblog.Dump(blog.Post(time.Now().Format("投稿 2006-01-02 15:04:05"), "本文を書く"))
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
