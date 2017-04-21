package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "os"
  "os/user"
  "path/filepath"
  "time"

  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/calendar/v3"

  "github.com/ajesler/playbulb-candle"
)

func BookEvent(srv *calendar.Service) (*calendar.Event, error) {
  // TODO: Use the current DateTime.
  // TODO: Set the correct room.
  // TODO: Set the correct calendar ID.
  timezone := "Pacific/Auckland"

  event := &calendar.Event {
    Summary: "A Meeting",
    Location: "This pod",
    Description: "Booked using a Candle!",
    Start: &calendar.EventDateTime {
      DateTime: time.Now().Format(time.RFC3339),
      TimeZone: timezone,
    },
    End: &calendar.EventDateTime {
      DateTime: time.Now().Add(time.Minute * 30).Format(time.RFC3339),
      TimeZone: timezone,
    },
  }

  calendarId := "primary"
  return srv.Events.Insert(calendarId, event).Do()
}

// based on https://developers.google.com/google-apps/calendar/quickstart/go

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
  cacheFile, err := tokenCacheFile()
  if err != nil {
    log.Fatalf("Unable to get path to cached credential file. %v", err)
  }
  tok, err := tokenFromFile(cacheFile)
  if err != nil {
    tok = getTokenFromWeb(config)
    saveToken(cacheFile, tok)
  }
  return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  fmt.Printf("Go to the following link in your browser then type the "+
    "authorization code: \n%v\n", authURL)

  var code string
  if _, err := fmt.Scan(&code); err != nil {
    log.Fatalf("Unable to read authorization code %v", err)
  }

  tok, err := config.Exchange(oauth2.NoContext, code)
  if err != nil {
    log.Fatalf("Unable to retrieve token from web %v", err)
  }
  return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
  usr, err := user.Current()
  if err != nil {
    return "", err
  }
  tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
  os.MkdirAll(tokenCacheDir, 0700)
  return filepath.Join(tokenCacheDir,
    url.QueryEscape("calendar-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
  f, err := os.Open(file)
  if err != nil {
    return nil, err
  }
  t := &oauth2.Token{}
  err = json.NewDecoder(f).Decode(t)
  defer f.Close()
  return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
  fmt.Printf("Saving credential file to: %s\n", file)
  f, err := os.Create(file)
  if err != nil {
    log.Fatalf("Unable to cache oauth token: %v", err)
  }
  defer f.Close()
  json.NewEncoder(f).Encode(token)
}

func nextEvents(srv *calendar.Service) []*calendar.Event {
  t := time.Now().Format(time.RFC3339)

  events, err := srv.Events.List("primary").ShowDeleted(false).
    SingleEvents(true).TimeMin(t).MaxResults(3).OrderBy("startTime").Do()

  if err != nil {
    log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
  }

  return events.Items
}

func effectFromEvent(e *calendar.Event, canBookNext bool) *playbulb.Effect {
  now := time.Now()
  t, _ := time.Parse(time.RFC3339, e.Start.DateTime)
  delta := t.Sub(now)
  fmt.Println("delta is ", delta)

  switch {
    case delta > (time.Hour * 2):
      fmt.Println("> 2 hours")
      c, _ := playbulb.ColourFromHexString("00000000")
      return playbulb.NewEffect(playbulb.SOLID, c, 0)
    case delta > (time.Minute * 20):
      fmt.Println("> 20 minutes")
      c, _ := playbulb.ColourFromHexString("0000FFFF")
      return playbulb.NewEffect(playbulb.SOLID, c, 0)
    case delta > (time.Minute * 10):
      fmt.Println("> 10 minutes")
      c, _ := playbulb.ColourFromHexString("000000FF")
      return playbulb.NewEffect(playbulb.SOLID, c, 0)
    default:
      fmt.Println("default effect")
      return defaultEffect
  }
}

func currentEvent(es []*calendar.Event) *calendar.Event {
  now := time.Now()

  for _, e := range es {
    if e.Start.DateTime != "" && e.End.DateTime != "" {
      st, _ := time.Parse(time.RFC3339, e.Start.DateTime)
      et, _ := time.Parse(time.RFC3339, e.End.DateTime)
      if st.Unix() < now.Unix() && now.Unix() < et.Unix() {
        return e
      }
    }
  }

  return nil
}

func nextEvent(es []*calendar.Event) *calendar.Event {
  if len(es) > 0 {
    return es[0]
  } else {
    return nil
  }
}

func canBookNextSlot(curEvent, nextEvent *calendar.Event) bool {
  if nextEvent == nil {
    return true
  }

  if curEvent.End.DateTime != "" && nextEvent.Start.DateTime != "" {
    cet, _ := time.Parse(time.RFC3339, curEvent.End.DateTime)
    nst, _ := time.Parse(time.RFC3339, nextEvent.Start.DateTime)

    return (nst.Sub(cet)) >= (time.Minute * 30)
  }

  return false
}

func futureEvents(es []*calendar.Event) []*calendar.Event {
  now := time.Now()

  fE := make([]*calendar.Event, 0)

  for _, e := range es {
    if e.Start.DateTime != "" {
      t, _ := time.Parse(time.RFC3339, e.Start.DateTime)
      if t.Unix() > now.Unix() {
        fE = append(fE, e)
      }
    }
  }

  return fE
}

var (
  done = make(chan bool, 1)
  defaultColour, _ = playbulb.ColourFromHexString("00FF00FF")
  defaultEffect = playbulb.NewEffect(playbulb.SOLID, defaultColour, 10)
)

func main() {
  ctx := context.Background()

  b, err := ioutil.ReadFile("client_secret.json")
  if err != nil {
    log.Fatalf("Unable to read client secret file: %v", err)
  }

  // If modifying these scopes, delete your previously saved credentials
  // at ~/.credentials/calendar-go-quickstart.json
  config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
  if err != nil {
    log.Fatalf("Unable to parse client secret file to config: %v", err)
  }
  client := getClient(ctx, config)

  srv, err := calendar.New(client)
  if err != nil {
    log.Fatalf("Unable to retrieve calendar Client %v", err)
  }

  candle := playbulb.NewCandle("e1817cd1d2cd4c088a094b1c31223588")
  err = candle.Connect()
  if err != nil {
    fmt.Println("Connection error:", err)
  }

  defer candle.Disconnect()

  for {
    events := nextEvents(srv)
    fEvents := futureEvents(events)

    fmt.Println("Upcoming events:")
    if len(fEvents) > 0 {
      for _, i := range fEvents {
        var when string
        // If the DateTime is an empty string the Event is an all-day Event.
        // So only Date is available.
        if i.Start.DateTime != "" {
          when = i.Start.DateTime
        } else {
          when = i.Start.Date
        }
        fmt.Printf("%s (%s)\n", i.Summary, when)
      }

      curEvent := currentEvent(events)
      nEvent := nextEvent(fEvents)

      canBook := canBookNextSlot(curEvent, nEvent)

      effect := effectFromEvent(nEvent, canBook)
      fmt.Println("Setting effect to", effect)
      candle.SetEffect(effect)
    } else {
      fmt.Printf("No upcoming events found.\n")
    }

    <-time.After(time.Minute * 1)
  }
}