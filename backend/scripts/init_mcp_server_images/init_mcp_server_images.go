package main

import (
	"fmt"

	. "github.com/AbhinavPalacharla/xtrn-personal/internal/mcp-server-images"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
)

func main() {
	googleCalendarImage, err := NewGoogleCalendarImage()

	if err != nil {
		ErrorLogger.Fatal(err)
	} else {
		fmt.Print("✅ Created Google Calendar Image\n")
	}

	airBNBImage, err := NewAirBNBImage()

	if err != nil {
		ErrorLogger.Fatal(err)
	} else {
		fmt.Print("✅ Created AirBNB Image\n")
	}

	_ = googleCalendarImage
	_ = airBNBImage
}
