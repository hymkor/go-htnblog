package htnblog

import (
	"encoding/xml"
	"testing"
)

var feedSample = []byte(`<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom"
      xmlns:app="http://www.w3.org/2007/app">
  <link rel="first" href="https://blog.hatena.ne.jp/{はてなID}}/{ブログID}/atom/entry" />
  <link rel="next" href="https://blog.hatena.ne.jp/{はてなID}/{ブログID}/atom/entry?page=1377584217" />
  <title>ブログタイトル</title>
  <link rel="alternate" href="http://{ルートURL}/"/>
  <updated>2013-08-27T15:17:06+09:00</updated>
  <author>
    <name>{はてなID}</name>
  </author>
  <generator uri="http://blog.hatena.ne.jp/" version="100000000">Hatena::Blog</generator>
  <id>hatenablog://blog/2000000000000</id>

  <entry>
    <id>tag:blog.hatena.ne.jp,2013:blog-{はてなID}-20000000000000-3000000000000000</id>
    <link rel="edit" href="https://blog.hatena.ne.jp/{はてなID}/{ブログID}/atom/entry/2500000000"/>
    <link rel="alternate" type="text/html" href="http://{ルートURL}/entry/2013/09/02/112823"/>
    <author><name>{はてなID}</name></author>
    <title>記事タイトル</title>
    <updated>2013-09-02T11:28:23+09:00</updated>
    <published>2013-09-02T11:28:23+09:00</published>
    <app:edited>2013-09-02T11:28:23+09:00</app:edited>
    <summary type="text"> 記事本文 リスト1 リスト2 内容 </summary>
    <content type="text/x-hatena-syntax">
      ** 記事本文
      - リスト1
      - リスト2
      内容
    </content>
    <hatena:formatted-content type="text/html" xmlns:hatena="http://www.hatena.ne.jp/info/xmlns#">
      <div class="section">
      <h4>記事本文</h4>
      <ul>
      <li>リスト1</li>
      <li>リスト2</li>
      </ul><p>内容</p>
      </div>
    </hatena:formatted-content>
    <app:control>
      <app:draft>no</app:draft>
      <app:preview>no</app:preview>
    </app:control>
  </entry>
  <entry>
  ...
  </entry>
  ...
</feed>`)

func TestFeed(t *testing.T) {
	var feed xmlFeed

	err := xml.Unmarshal(feedSample, &feed)
	if err != nil {
		t.Fatal(err.Error())
	}
	nextUrl := feed.nextUrl()
	expect := "https://blog.hatena.ne.jp/{はてなID}/{ブログID}/atom/entry?page=1377584217"
	if nextUrl != expect {
		t.Fatalf("expect '%s' but '%s'", expect, nextUrl)
	}

	editUrl := feed.Entry[0].EditUrl()
	expect = "https://blog.hatena.ne.jp/{はてなID}/{ブログID}/atom/entry/2500000000"

	if editUrl != expect {
		t.Fatalf("expect '%s' but '%s'", expect, editUrl)
	}
}
