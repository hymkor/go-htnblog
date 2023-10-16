go-htnblog：はてなブログ投稿用Go言語パッケージ
==============================================

まだ BASIC 認証でごめん

サンプルツール
-----------------------

[cmd/htnblog/main.go](cmd/htnblog/main.go)

```cmd/htnblog/htnblog |
Usage: htnblog {list|new|edit}
  htnblog list ... show recent articles
  htnblog new  ... create new draft
  htnblog edit ... edit the latest article
    The lines in the draft up to "---" are the header lines,
    and the rest is the article body.

Please write your setting on ~/.htnblog as below:
    {
        "userid":"(YOUR_USER_ID)",
        "endpointurl":"(END_POINT_URL)",
        "apikey":"(YOUR API KEY)",
        "author":"(YOUR NAME)",
        "editor":"(YOUR EDITOR.THIS IS for cmd/htnblog/main.go)"
    }
```

例: 一覧表示
------------

[examples/list.go](examples/list.go)

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

例: 新規投稿
------------

[examples/post.go](examples/post.go)

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

例: 最も新しい記事を編集
------------------------

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

設定ファイル例
--------------

[sample.json](sample.json)

```sample.json
{
    "userid":"(YOUR_USER_ID)",
    "endpointurl":"(END_POINT_URL)",
    "apikey":"(YOUR API KEY)",
    "author":"(YOUR NAME)",
    "editor":"(YOUR EDITOR.THIS IS for cmd/htnblog/main.go)"
}
```
