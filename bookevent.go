package candlecalendar

import (
  "time"
  "google.golang.org/api/calendar/v3"
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