package main

import (
    "encoding/json"
    "net/http"
	"bytes"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
    "fmt"
    "os"
    "io/ioutil"
)

type Button struct {
    Name    string
    Url     string
    Mac     string
}
type Configuration struct {
    Buttons []Button
    Nic     string
}

func loadConfig() Configuration {
    file, _ := os.Open("conf.json")
    decoder := json.NewDecoder(file)
    configuration := Configuration{}
    err := decoder.Decode(&configuration)
    if err != nil {
        fmt.Println("error:", err)
    }
    fmt.Println(configuration.Buttons)
    return configuration
}


func makeRequest(url string) {
	res, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", body)
}


func main() {

    var configuration = loadConfig()

	log.Printf("Starting up on interface[%v]...", configuration.Nic)

	h, err := pcap.OpenLive(configuration.Nic, 65536, true, pcap.BlockForever)

	if err != nil || h == nil {
		log.Fatalf("Error opening interface: %s\nPerhaps you need to run as root?\n", err)
	}
	defer h.Close()

    var filter = "arp and ("
    for _,button := range configuration.Buttons {
        mac, err := net.ParseMAC(button.Mac)
        if err != nil {
            log.Fatal(err)
        }
        filter += "(ether src host " + mac.String() + ")"
    }
    filter += ")"

	//err = h.SetBPFFilter("arp and ((ether src host " + gatoradeDashButton.String() + ") or (ether src host " + gladDashButton.String() + "))")
    err = h.SetBPFFilter(filter)
	if err != nil {
		log.Fatalf("Unable to set filter! %s\n", err)
	}
	log.Println("Listening for Dash buttons...")

	packetSource := gopacket.NewPacketSource(h, h.LinkType())

	// Since we're using a BPF filter to limit packets to only our buttons, we don't need to worry about anything besides MAC here...
	for packet := range packetSource.Packets() {
		ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
         for _,button := range configuration.Buttons {
            mac, err := net.ParseMAC(button.Mac)
            if err != nil {
                log.Fatal(err)
            }
		    if bytes.Equal(ethernetPacket.SrcMAC, mac) {
                log.Printf("Button [%v] was pressed.", button.Name)
                makeRequest(button.Url)
            } else {
			    log.Printf("Received button press, but don't know how to handle MAC[%v]", ethernetPacket.SrcMAC)
            }
		}
	}

}