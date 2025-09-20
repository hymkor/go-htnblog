htnblog-go：はてなブログ-クライアント
====================================

[![GoDev](https://pkg.go.dev/badge/github.com/hymkor/htnblog-go)](https://pkg.go.dev/github.com/hymkor/htnblog-go)

本パッケージは

- コマンドラインクライアント(htnblog.exe, htnblog)  
    → シェル、コマンドプロンプトなどから利用。vim など任意のテキストエディターで記事を編集して、投稿
- クライアントライブラリ(github.com/hymkor/htnblog-go)  
    → Go言語で投稿機能を利用できるライブラリ

からなります。Windows, Linux (Ubuntu on WSL) で動作を確認しています

※ 2025.09.17 GitHubレポジトリ名、Goのパッケージ名を `go-htnblog` から`htnblog-go` へ変更しましたが、実行ファイル名ほかは特に変更ありません。ライブラリとしての利用はほとんどなく、主に実行ファイルの利用がメインになっているのに、名前の頭に `go-` がのっかってるのがすわりが悪くて、変えてしまいました。


コマンドラインクライアント
--------------------------

### 手作業でインストール

[Releases](https://github.com/hymkor/htnblog-go/releases) よりダウンロードして、環境変数 PATH で指定されたディレクトリに実行ファイル htnblog.exe を展開してください

#### Go一式がある場合

```
go install github.com/hymkor/htnblog-go/cmd/htnblog@latest
```

#### scoop インストーラーがある場合

```
scoop install https://raw.githubusercontent.com/hymkor/htnblog-go/master/htnblog.json
```

もしくは

```
scoop bucket add hymkor https://github.com/hymkor/scoop-bucket
scoop install htnblog
```


### 使い方

[cmd/htnblog/main.go](https://github.com/hymkor/htnblog-go/blob/master/cmd/htnblog/main.go)

#### htnblog (オプションなし)

ヘルプを表示します

```./htnblog |
htnblog v1.2.0-6-gc339b6a-windows-amd64 by go1.25.1

Usage: htnblog {options...} {init|list|new|type|edit}
  htnblog init                   ... edit configuration
  htnblog list                   ... show recent articles
  htnblog new                    ... create a new draft
  htnblog type    {URL|@0|@1|..} ... output the article to STDOUT
  htnblog edit    {URL|@0|@1|..} ... edit the article
  htnblog delete  {URL|@0|@1|..} ... output the article to STDOUT and delete it
  htnblog publish {URL|@0|@1|..} ... set false the draft flag of the article
  htnblog browse  {URL|@0|@1|..} ... open the edit page in a web browser
  htnblog export                 ... export all articles
                                       as "./entry-YYYY-MM-DD-NNNNNN.md"

    The lines in the draft up to "```" are the header lines,
    and the rest is the article body.

Usage of ./htnblog:
  -1	Use the value of "endpointurl1" in the JSON setting
  -2	Use the value of "endpointurl2" in the JSON setting
  -3	Use the value of "endpointurl3" in the JSON setting
  -debug
    	Enable Debug Output
  -f	Delete without prompt
  -n int
    	fetch articles (default 10)
  -rc string
    	use the specified file instead of ~/.htnblog
  -header-start string
    	header start line
  -header-end string
    	header end line
```

#### htnblog init

設定を編集します。

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

```json
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

#### htnblog list

直近10件の記事のリストを表示します。

#### htnblog new

新規記事のドラフトを作成します。テキストエディターが起動します。

    ```header
    Rem: EndPointUrl: https://blog.hatena.ne.jp/zetamatta/zetamatta.hatenablog.com/atom
    Category: 
    Title: 
    ```

`` ```header `` から `` ``` `` まではヘッダ行になります。記事の設定情報が編集可能です。

- `Category:` は、はてなブログのカテゴリ名を空白で区切ってならべます。
- `Title:` 以下には記事のタイトルを記述します。
- `Rem:` で始まる行は参考情報です。変更しても反映されず、無視されます。
- 本文の markdown はヘッダ行の後: 二番目の `` ``` `` 以降に記述してください。
- ヘッダの開始行、終了行は -header-start , -header-end で変更可能です

エディターを終了すると、`Are you sure to post ? (y/n/edit):` というプロンプトが表示されます。`y` で投稿実行、`n` はドラフトを破棄、`e` で編集しなおしとなります。

#### htnblog edit {URL|@0|@1|…}

既存記事の編集。記事を指定しない場合は @0 = 最も最近に編集したページを対象とします。
`htnblog new` と同様にテキストエディターが起動します。

#### htnblog publish {URL|@0|@1|…}

指定した記事を下書き状態から公開状態へ変更します。

#### htnblog browse {URL|@0|@1|…}

指定した記事の編集ページをウェブブラウザで開きます。

#### htnblog export

全記事をカレントディレクトリに entry-YYYY-MM-DD-NNNNNN.md 形式でバックアップします。

Goライブラリ htnblog-go
------------------------

### 使用例: 一覧表示

[examples/list.go](https://github.com/hymkor/htnblog-go/blob/master/examples/list.go)

設定は標準入力から読み込むようにしてます。ブログに投稿するためのアカウント情報を保持する htnblog.Blog型のインスタンスは `htnblog.NewFromJSON` 関数で JSON テキストから生成していますが、別に `&Blog{...}` でいきなり生成しても構いません。

```examples/list.go
package main

import (
    "fmt"
    "io"
    "os"

    "github.com/hymkor/htnblog-go"
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

[examples/post.go](https://github.com/hymkor/htnblog-go/blob/master/examples/post.go)

`(*Blog) Post` 関数は戻り値としてウェブアクセスの返信を `*http.Response` と `error` で返してきます。エラーでない場合は `*https.Response` の Body という io.Reader を全部読みとって、Close しなければいけないのですが、それらはエラーの判断も含めて `(*Blog).DropResponse` というメソッドに委ねています。同メソッドでは通常は読み取ったデータを io.Discard に捨てていますが、DebugPrint という io.Writer のフィールドが nil でない場合はそちらへコピーします
( つまり例では標準エラー出力にサーバーのレスポンスを表示させています )

`$ go run examples/post.go < ~/.htnblog`

```examples/post.go
package main

import (
    "fmt"
    "io"
    "os"
    "time"

    "github.com/hymkor/htnblog-go"
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
    blog.DebugPrint = os.Stderr
    return blog.DropResponse(blog.Post(time.Now().Format("投稿 2006-01-02 15:04:05"), "本文を書く", "yes"))
}

func main() {
    if err := post(); err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        os.Exit(1)
    }
}
```

### 使用例: 最も新しい記事を編集

[examples/edit.go](https://github.com/hymkor/htnblog-go/blob/master/examples/edit.go)

`$ go run examples/edit.go < ~/.htnblog`

```examples/edit.go
package main

import (
    "errors"
    "fmt"
    "io"
    "os"
    "time"

    "github.com/hymkor/htnblog-go"
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
    blog.DebugPrint = os.Stderr
    return blog.DropResponse(blog.Update(entries[0]))
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
