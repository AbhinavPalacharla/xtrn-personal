package main

import (
	"fmt"

	mcp_server_images "github.com/AbhinavPalacharla/xtrn-personal/internal/mcp-server-images"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
)

func main() {
	googleCalendarImage, err := mcp_server_images.NewGoogleCalendarImage()

	if err != nil {
		StdErrLogger.Fatal(err)
	} else {
		fmt.Print("✅ Created Google Calendar Image\n")
	}

	gmailImage, err := mcp_server_images.NewGmailImage()

	if err != nil {
		StdErrLogger.Fatal(err)
	} else {
		fmt.Print("✅ Created Gmail Image\n")
	}

	airBNBImage, err := mcp_server_images.NewAirBNBImage()

	if err != nil {
		StdErrLogger.Fatal(err)
	} else {
		fmt.Print("✅ Created AirBNB Image\n")
	}

	_ = googleCalendarImage
	_ = gmailImage
	_ = airBNBImage
}
