package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	iterator "google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
    firebaseConfigFile = "secrets/firebaseConfig.json"
)

var (
    ctx context.Context
    app *firebase.App
    client *firestore.Client
)

type Record struct {
    Timestamp time.Time
    Duration  int
}

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

type Cookie struct {
    Name    string
    Value   string

    Path    string
    Expires time.Time
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
        <h1 hx-post="/game" hx-swap="outerHTML ignoreTitle:true transition:true">
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
        <input name="answer" hx-trigger="keyup[keyCode==13]" hx-target="#wrapped" hx-get="/action?punchline={{.Punchline}}" autofocus></input>
        <div id="wrapped">
            <h1 id="banner"></h1>
            <div class="hidden">
                <h1>you got the joke in <br></br> <span id="blue">2 minutes</span> and <span id="green">31 seconds</span></h1>
                <h3>you beat <span id="yellow">25</span>% of people</h3>
            </div>
        </div>
    </div>
</html>
`

const tpl3 = `
<div id="wrapped">
    <h1 id="banner">{{.Banner}}</h1>
    <div {{.Win}} class="hidden">
        <h1>you got the joke in <br></br> <span id="blue">{{.Minutes}} minutes</span> and <span id="green">{{.Seconds}} seconds</span></h1>
        <h3>you beat <span id="yellow">{{.Percent}}</span>% of people</h3>
    </div>
</div>
`

func GetRecords() []Record {
    var records []Record
    iter := client.Collection("wins").Documents(ctx)
    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            log.Fatalf("Failed to iterate: %v", err)
        }
        var record Record
        doc.DataTo(&record)
        records = append(records, record)
        log.Println(records)
        
        //log.Println(doc.Data())
    }

    return records
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, tpl)
}

func handler2(w http.ResponseWriter, r *http.Request) {
    prevCookie, _ := r.Cookie("startTime")
    if prevCookie == nil {
        now := time.Now().UTC()
        midnight := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.UTC().Location())
        seconds := midnight.Sub(now).Seconds()
        cookie := http.Cookie{
            Name:    "startTime",
            Value:   time.Now().UTC().String(),
            Path:    "/",
            Expires: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.UTC().Location()),
            SameSite: http.SameSiteLaxMode,
            MaxAge: int(seconds),
        }

        http.SetCookie(w, &cookie)
    }

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
    var diff float64 = 0.0
    var finalSeconds int = 0
    var finalMinutes int = 0
    params := r.URL.Query()
    w.Header().Set("Content-Type", "text/html")
    t, _ :=  template.New("webpage").Parse(tpl3)
    banner := ""
    var win template.HTMLAttr
    var percent float64
    percent = 100

    if strings.ToLower(params.Get("answer")) == params.Get("punchline") {
        cookie, err := r.Cookie("startTime")
        if err != nil {
            switch {
            case errors.Is(err, http.ErrNoCookie):
                http.Error(w, "cookie not found", http.StatusBadRequest)
            default:
                log.Println(err)
                http.Error(w, "server error", http.StatusInternalServerError)
            }
            return
        } else {
            endCookie , _ := r.Cookie("endTime")
            if endCookie == nil {
                start, _ := time.Parse("2006-01-02 15:04:05.999999 Z0700 MST", cookie.Value)
                now := time.Now().UTC()

                diff = now.Sub(start).Seconds()
                finalMinutes = int(diff) / 60
                finalSeconds = int(diff) % 60

                _, _, _ = client.Collection("wins").Add(ctx, map[string]interface{}{
                    "duration": int(diff),
                    "timestamp": now,
                })

                midnight := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.UTC().Location())
                tonight := midnight.Sub(now).Seconds()
                cookie2 := http.Cookie{
                    Name:    "endTime",
                    Value:   now.String(),
                    Path:    "/",
                    Expires: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.UTC().Location()),
                    SameSite: http.SameSiteLaxMode,
                    MaxAge: int(tonight),
                }
                
                http.SetCookie(w, &cookie2)
            } else {
                start, _ := time.Parse("2006-01-02 15:04:05.999999 Z0700 MST", cookie.Value)
                endTime, _ := time.Parse("2006-01-02 15:04:05.999999 Z0700 MST", endCookie.Value)
                diff = endTime.Sub(start).Seconds()
                finalMinutes = int(diff) / 60
                finalSeconds = int(diff) % 60
            }
        }

        //banner = fmt.Sprintf("you're so punny! You took %.2f seconds to finish", diff)
        win = "id=\"popup\""
        now := time.Now().UTC()
        startOfDay := time.Date(now.Year(),now.Month(), now.Day(), 0, 0, 0 ,0, now.UTC().Location())
        midnight := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.UTC().Location())
        recordCount := 0
        records := GetRecords()
        better := 0
        for _, record := range records {
            log.Println(record.Duration)
            log.Println(startOfDay)
            log.Println(midnight)
            log.Println(record.Timestamp)
            log.Println(startOfDay.After(record.Timestamp))
            log.Println(midnight.Before(record.Timestamp))
            if startOfDay.Before(record.Timestamp) && midnight.After(record.Timestamp) {
                recordCount++
                if int(diff) > record.Duration {
                    better++
                }
            }
        }
        difference := recordCount - better
        if recordCount > 0 {
            percent = (float64(difference) / float64(recordCount)) * 100
        } else {
            percent = 100
        }

    } else {
        banner = "*crickets*"
    }

    data := struct {
        Percent int
        Minutes int 
        Seconds int 
        Banner  string
        Win     template.HTMLAttr
    }{
        Minutes: finalMinutes,
        Seconds: finalSeconds,
        Banner: banner,
        Win: win,
        Percent: int(percent),
    }
    _ = t.Execute(w, data)
}

func handler4(w http.ResponseWriter, r *http.Request) {
    var records []Record
    iter := client.Collection("wins").Documents(ctx)
    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            log.Fatalf("Failed to iterate: %v", err)
        }
        var record Record
        doc.DataTo(&record)
        records = append(records, record)
        log.Println(records)
        
        //log.Println(doc.Data())
    }
}

func main() {
    port := os.Getenv("PORT")
    
    ctx = context.Background()
    opt := option.WithCredentialsFile(firebaseConfigFile)
    app, err := firebase.NewApp(ctx, nil, opt)
    if err != nil {
        log.Fatalf("Firebase initialization error: %v\n", err)
    }

    client, err = app.Firestore(ctx)
    if err != nil {
        log.Fatalf("Firestore initalization error: %v\n", err)
    }

    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    http.HandleFunc("/", handler)
    http.HandleFunc("/game", handler2)
    http.HandleFunc("/action", handler3)
    http.HandleFunc("/firestore", handler4)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
