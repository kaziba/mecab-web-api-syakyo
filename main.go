package main

import (
  "bitbucket.org/rerofumi/mecab"
  "bufio"
  "encoding/json"
  "flag"
  "fmt"
  "io"
  "net/http"
  "os"
  "strconv"
)

type MecabOutput struct {
  Status int
  Code   string
  Result []string
}

type MecabInput struct {
  Token    string
  Sentence string
}

func apiRequest(w http.ResponseWriter, r *http.Request) {

  // 配列を定義する。
  // 配列は固定長、スライスは可変長の配列のようなもの。理由がない限りはスライスを使えばよい。
  // http://ashitani.jp/golangtips/tips_slice.html
  list := make([]string, 0)

  ret := MecabOutput{0, "OK", list}
  request := ""

  /**
   * JSON return
   */
  // deferステートメントは、deferを実行した関数がリターンする直前に、指定した関数の呼び出しが行われるようにスケジューリングします。
  // このコードでいえば、なんやかんやしたあと、最後に呼ばれる
  // http://golang.jp/effective_go#defer
  defer func() {
    // json.Marshalは構造体からJSON文字列への変換する関数
    // JSON形式でレスポンスを返すために変換している。
    outjson, err := json.Marshal(ret)
    if err != nil {
      fmt.Println(err)
    }
    w.Header().Set("Content-Type", "application/json")
    // w.WriteHeader(400) <- 200のときは不要。
    fmt.Fprint(w, string(outjson)) // res.send() みたいなもん
  }()

  // type check
  if r.Method != "POST" {
    ret.Status = 1
    ret.Code = "Not POST method"
    return
  }

  /**
   * request body
   */
  fmt.Println(r)
  // => &{POST / HTTP/1.1 1 1 map[Accept-Language:[ja,en-US;q=0.8,en;q=0.6] Cookie:[REVEL_FLASH=] Content-Length:[76] Origin:[chrome-extension://aejoelaoggembcahagimdiliamlcdmfm] User-Agent:[Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.87 Safari/537.36] Content-Type:[application/json] Accept:[*/*] Accept-Encoding:[gzip, deflate] Connection:[keep-alive]] 0xc820014540 76 [] false 192.168.33.10:9000 map[] map[] <nil> map[] 192.168.33.1:55534 / <nil> <nil>}

  // fmt.Println(r.Body)
  // =>  解読不能

  rb := bufio.NewReader(r.Body)
  // fmt.Println(rb)
  // => 解読不能

  for {
    s, err := rb.ReadString('\n')
    fmt.Println(s)
    /**
            {

      		     "Token": "01234",

      		     "Sentence": "すもももももももものうち"

      		  }
    */
    request = request + s
    if err == io.EOF {
      break
    }
  }

  /**
   * JSON parse
   */
  var dec MecabInput
  b := []byte(request)

  // json.Unmarshalは、構造体のjsonタグがあればその値を対応するフィールドにマッピングする
  // (ex)
  // var mt MyType
  // json.Unarshal([]byte(`{"A":"aaa", "FooBar":"baz"}`, &mt)
  // fmt.Printf("%#v\n", mt)
  err := json.Unmarshal(b, &dec)

  fmt.Println(b)
  // => [123 10 32 32 34 84 111 107 101 110 34 58 32 34 48 49 50 51 52 34 44 10 32 32 34 83 101 110 116 101 110 99 101 34 58 32 34 227 129 153 227 130 130 227 130 130 227 130 130 227 130 130 227 130 130 227 130 130 227 130 130 227 130 130 227 129 174 227 129 134 227 129 161 34 10 125]

  fmt.Println(&dec)
  // => &{01234 すもももももももものうち}

  if err != nil {
    ret.Status = 2
    ret.Code = "JSON parse error."
    return
  }

  /**
   * mecab parse
   */
  result, err := mecab.Parse(dec.Sentence)
  if err == nil {
    fmt.Println(result)
    /**
      [すもも 名詞,一般,*,*,*,*,すもも,スモモ,スモモ も       助詞,係助詞,*,*,*,*,も,モ,モ も
       も      名詞,一般,*,*,*,*,もも,モモ,モモ も     助詞,係助詞,*,*,*,*,も,モ,モ もも
       名詞,一般,*,*,*,*,もも,モモ,モモ の     助詞,連体化,*,*,*,*,の,ノ,ノ うち       名詞,非
       自立,副詞可能,*,*,*,うち,ウチ,ウチ]
    */
    // for i, n := range result { ← これだとfor文中でiを用いる処理がないためエラーが発生する。(Lintの関係？) その代わりの_
    for _, n := range result {
      fmt.Println(n)
      ret.Result = append(ret.Result, n)
    }
  }
}

func main() {

  // コマンドラインオプションの解析
  var portNum int
  flag.IntVar(&portNum, "port", 80, "int flag")
  flag.IntVar(&portNum, "p", 80, "int flag")
  flag.Parse()

  var port string
  port = ":" + strconv.Itoa(portNum) // strconv.Itoa() 数値 -> 文字列
  fmt.Println("listen port = ", port)

  http.HandleFunc("/", apiRequest)

  err := http.ListenAndServe(port, nil)

  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
