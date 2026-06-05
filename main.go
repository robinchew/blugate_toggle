package main

import (
	"encoding/hex"
	"fmt"
	"log/slog"

	"os"
	"tinygo.org/x/bluetooth"
)

// Replace these with your target device's UUIDs
/*
var (
	targetDeviceName = "MyBLEDevice"
	serviceUUID      = bluetooth.ServiceUUIDPeripheralChannel // Or a specific UUID
	charUUID         = bluetooth.NewUUID([16]byte{0x00, 0x00, 0x2a, 0x37, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0x80, 0x5f, 0x9b, 0x34, 0xfb}) // Example: Heart Rate Measurement
)
*/
var (
	targetDeviceName = "nimble-ble"
	foundDevice bluetooth.ScanResult
)

func main() {
	adapter := bluetooth.DefaultAdapter
	slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Set the minimum level to DEBUG to include debug logs, or INFO
    slog.SetLogLoggerLevel(slog.LevelDebug)

	// 1. Enable the BLE adapter
	must(adapter.Enable())
		// 2. Scan for the device
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		slog.Debug("Found Device \n", "local_name", result.LocalName(), "address", result.Address.String(), "RSSI", result.RSSI)
		if result.Address.String() == "58:8C:81:9F:B7:A2" {
			foundDevice = result
			adapter.StopScan()
		}
	})
	must(err)

	// Set the handler before connecting or advertising
    adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
        if connected {
            fmt.Println("device connected:", device.Address.String())
        } else {
            fmt.Println("device DISCONNECTED:", device.Address.String())
            // Handle disconnect logic here
        }
    })

	device, err := adapter.Connect(foundDevice.Address, bluetooth.ConnectionParams{})
	must(err)
	defer device.Disconnect()

	serviceUUID, _ := bluetooth.ParseUUID("b2bbc642-46da-11ed-b878-0242ac120002")
	charUUID, _ := bluetooth.ParseUUID("9f0921bf-c468-46bd-a724-b95bfa95541e")

	// 4. Discover Services
	services, err := device.DiscoverServices([]bluetooth.UUID{serviceUUID}) // nil gets all services
	must(err)

	for _, service := range services {
		// 5. Discover Characteristics for each service
		chars, err := service.DiscoverCharacteristics([]bluetooth.UUID{charUUID})
		must(err)

		for _, char := range chars {
			// 6. Check if this is the characteristic we want
			// In this example, we enable notifications for ALL characteristics that support it
			// or you can filter by char.UUID()
			slog.Debug("Found Characteristic\n", "characteristic", char.String())

			err := char.EnableNotifications(func(buf []byte) {
				slog.Debug("Notification received\n", "characteristic", char.String(), "message_buf", buf, "message_hex", hex.EncodeToString(buf))
				fmt.Printf("%s\n", buf)
			})
			if err != nil {
				slog.Debug("Could not enable notifications\n", "characteristic", char.String(), "error", err.Error())
			} else {
				fmt.Printf("Receiving...\n")
				slog.Debug("Enabled notifications for characteristic\n", "characteristic", char.String())
			}
		}
	}

	// Keep the program running to receive notifications
	select {}
}

func must(err error) {
	if err != nil {
		slog.Error("Fatal Exit", "error", err)
		os.Exit(1)
	}
}
