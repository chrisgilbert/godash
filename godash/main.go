package main

// Made it use digest auth so I can trigger actions in Indigo. -dnewhall

import (
	"encoding/json"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
)

// Button is a Dash, from Amazon
type Button struct {
	Name     string
	URL      string
	Username string
	MAC      string
}

// Configuration is Network Interface and Buttons.
type Configuration struct {
	Buttons []Button
	NIC     string
}

func loadConfig() Configuration {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatalln("Error Loading Configuration:", err)
	}
	log.Println("Loaded", len(configuration.Buttons), "Button(s):")
	for _, button := range configuration.Buttons {
		log.Printf("- Button: %v (%v): %v\n", button.Name, button.MAC, button.URL)
	}
	return configuration
}

func main() {
	var configuration = loadConfig()
	log.Printf("Starting up on interface[%v]...", configuration.NIC)

	var filter = "arp and ("
	// Create a packet capture filter for the button's MAC addresses.
	for _, button := range configuration.Buttons {
		MAC, err := net.ParseMAC(button.MAC)
		if err != nil {
			log.Fatalf("Unable to parse MAC: %s (%s)\n", button.MAC, err)
		}
		if filter != "arp and (" {
			filter += " or "
		}
		filter += "(ether src host " + MAC.String() + ")"
	}
	filter += ")"
	capturePackates(configuration.NIC, filter, configuration.Buttons)
}

func capturePackates(NIC string, filter string, buttons []Button) {
	h, err := pcap.OpenLive(NIC, 65536, true, pcap.BlockForever)
	defer h.Close()
	if err != nil || h == nil {
		log.Fatalf("Error opening interface: %s\nPerhaps you need to run as root?\n", err)
	}

	if err = h.SetBPFFilter(filter); err != nil {
		log.Fatalf("Unable to set filter: %s %s\n", filter, err)
	}

	log.Println("Listening for Dash buttons...")
	packetSource := gopacket.NewPacketSource(h, h.LinkType())

	// Using a BPF filter to limit packets to only our buttons,
	// there is no need to capture anything besides MAC here.
	for packet := range packetSource.Packets() {
		ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
		for _, button := range buttons {
			if ethernetPacket.SrcMAC.String() == button.MAC {
				log.Println("Button", button.Name, "was pressed.")
				go makeRequest(button.URL, button.Username)
				break
			}
			log.Printf("Received button press, but don't know how to handle MAC[%v]\n", ethernetPacket.SrcMAC)
		}
	}
}

func makeRequest(url string, username string) {
	var cmd *exec.Cmd
	if username != "" {
		// Adding digest auth to Go looked like hell. This was a lot easier.
		cmd = exec.Command("curl", "-u", username, "--digest", url)
		cmd.Stderr = cmd.Stdout
		output, err := cmd.Output()
		if err != nil {
			log.Println("Error Curling URL", url, "->", err)
		} else {
			log.Println("Curl Output:", string(output))
		}
	} else {
		// TODO: don't hard code this to POST nor JSON. Put them in the config file.
		res, err := http.Post(url, "application/json", nil)
		if err != nil {
			log.Println("Error POSTing to URL", url, "->", err)
			return
		}
		defer func() {
			// This is how you win the game of `errcheck`.
			if err := res.Body.Close(); err != nil {
				log.Println("Failed to close HTTP response body:", err)
			}
		}()
		if output, err := ioutil.ReadAll(res.Body); err != nil {
			log.Println("Error POSTing to URL", url, "->", err)
		} else {
			log.Println("POST Output:", string(output))
		}
	}
}
