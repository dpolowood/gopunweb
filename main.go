package main

import (
    "log"
    "fmt"
    "math/rand/v2"
    "html/template"
    "net/http"
    "strings"
)

type Pun struct {
    Prompt    string
    Punchline string
    Full      string
}

var puns = []Pun{
    {
        "I used to be a baker, but I couldn't make enough _____", 
        "dough", 
        "I used to be a baker, but I couldn't make enough dough",
    },
    {
        "Why did the scarecrow win an award? Because he was ___________ in his field",
        "outstanding",
        "Why did the scarecrow win an award? Because he was outstanding in his field",
    },
    {
        "The past, the present, and the future walked into a bar. It was _____",
        "tense",
        "The past, the present, and the future walked into a bar. It was tense",
    },
    {
        "I wanted to learn how to drive a stick, but I couldn't find the ______",
        "manual",
        "I wanted to learn how to drive a stick, but I couldn't find the manual",
    },
    {
        "Why don't skeletons fight each other? They don't have the ____",
        "guts",
        "Why don't skeletons fight each other? They don't have the guts",
    },
}

const tpl = `
<!DOCTYPE html>
<html>
    <head>
        <script src="/static/htmx.min.js"></script>
        <link href="/static/style.css" rel="stylesheet" />
        <meta charset="UTF-8">
        <title>what's so punny?</title>
    </head>
    <body class="container">
        <h1 hx-post="/game" hx-swap="outerHTML ignoreTitle:true transition:true" hx-trigger="mouseenter">
            what's so punny?
        </h1>
    </body>
</html>
`

const tpl2 = `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
    </head>
    <div class="game">
        <h1>
            {{.Prompt}}
        </h1>
        <input name="answer" hx-trigger="keyup[keyCode==13]" hx-target="#banner" hx-get="/action?punchline={{.Punchline}}" autofocus></input>
        <h1 id="banner"></h1>
    </div>
</html>
`

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, tpl)
}

func handler2(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    random := rand.IntN(5)
    t, _ :=  template.New("webpage").Parse(tpl2)
    data := struct {
        Title     string
        Prompt    string
        Punchline string
    }{
        Title: "what's so punny",
        Prompt: puns[random].Prompt,
        Punchline: puns[random].Punchline,
    }

    _ = t.Execute(w, data)
}

func handler3(w http.ResponseWriter, r *http.Request) {
    params := r.URL.Query()
    keys := make([]string, len(params))
    i := 0
    for k := range params {
        keys[i] = k
        i++
    }
    banner := ""

    if strings.ToLower(params.Get("answer")) == params.Get("punchline") {
        banner = "you're so punny!"
    } else {
        banner = "*crickets*"
    }

    fmt.Fprintf(w, "%s", banner)
}

func main() {
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    http.HandleFunc("/", handler)
    http.HandleFunc("/game", handler2)
    http.HandleFunc("/action", handler3)
    log.Fatal(http.ListenAndServe(":443", nil))
}
