package main

import (
	"fmt"

	mcp_server_images "github.com/AbhinavPalacharla/xtrn-personal/internal/mcp-server-images"
	mcp_server_instances "github.com/AbhinavPalacharla/xtrn-personal/internal/mcp-server-instances"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
)

func main() {
	googleCalendarEnv := mcp_server_images.NewGoogleCalendarEnv("Qualcomm Calendar")
	googleCalendarInstance, err := mcp_server_instances.NewGoogleCalendarInstance(googleCalendarEnv)

	if err != nil {
		StdErrLogger.Fatal(fmt.Errorf("%w", err))
	} else {
		fmt.Printf("✅ Created Google Calendar Instance %v: 🚀%s\n", googleCalendarInstance.InstanceID, googleCalendarInstance.Address)
	}

	airbnbEnv := mcp_server_images.NewAirBNBEnv()
	airbnbInstance, err := mcp_server_instances.NewAirBNBInstance(airbnbEnv)

	if err != nil {
		StdErrLogger.Fatal(fmt.Errorf("%w", err))
	} else {
		fmt.Printf("✅ Created AirBNB Instance %v: 🚀%s\n", airbnbInstance, airbnbInstance.Address)
	}
}
