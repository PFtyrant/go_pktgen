package main
import (
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/pcap"
    cq "go_gen/queue"
    "flag"
    "log"
    "net"
    "time"
    "sync"
    "os"
    "os/signal"
    "syscall"
    "fmt"
)

var (
    device       string
    promiscuous  bool   = true
    err          error
    timeout      time.Duration = time.Nanosecond
    handle       *pcap.Handle
    buffer       gopacket.SerializeBuffer
    options      gopacket.SerializeOptions
    wg           sync.WaitGroup
    sigs         = make(chan os.Signal, 1)
    end          = make(chan struct{})
)

func init() {
    flag.StringVar(&device, "d", "a", "Device name")
    flag.Parse()
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
}

func nextIP(ip net.IP, inc uint) net.IP {
    i := ip.To4()
    v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
    v += inc
    v3 := byte(v & 0xFF)
    v2 := byte((v >> 8) & 0xFF)
    v1 := byte((v >> 16) & 0xFF)
    v0 := byte((v >> 24) & 0xFF)
    return net.IPv4(v0, v1, v2, v3)
}

func main() {
    chan_q := cq.NewChanQueue(1<<16)
    wg.Add(2)
    // Packet generator
    go func (CQ *cq.ChanQueue) {
        // packet custumize. It need to move.
        char_str := "1234567890abcdefghijklmnopqrstuvwxyzffff"
        res_str := ""
        raw_num := 5
        for i:= 0; i < raw_num; i++ {
            res_str = res_str + char_str
        }
        rawBytes := []byte(res_str)
        fmt.Println("length of rawbyte : ", len(rawBytes))
        // This time lets fill out some information
        ethernetLayer := &layers.Ethernet{
            SrcMAC: net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
            DstMAC: net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
            EthernetType: 0x0800,
        }
        udpLayer := &layers.UDP{
            Length : 8,
            SrcPort: layers.UDPPort(65535),
            DstPort: layers.UDPPort(65535),
        }
        ipLayer := &layers.IPv4{
            SrcIP: net.IP{172, 16, 0, 100},
            DstIP: net.IP{100, 0, 0, 0},
            Protocol: 17,
            TTL: 64,
            Version : 4,
            IHL : 5,
            Length : uint16(len(rawBytes)+20+8),
        }
        num := 1 << 24
        buffer = gopacket.NewSerializeBuffer()
        LOOP:
        for i := 0; i < num; i++ {
            select {
                case  <-sigs:
                    fmt.Println("sigs up")
                    break LOOP
                default:
                    gopacket.SerializeLayers(buffer, options,
                        ethernetLayer,
                        ipLayer,
                        udpLayer,
                        gopacket.Payload(rawBytes))
                    CQ.Push(buffer.Bytes())
                    buffer.Clear()
                    ipLayer.SrcIP = nextIP(ipLayer.SrcIP, 1)
                    continue
            }
        }
        close(end)
        wg.Done()
    }(chan_q)

    go transmitter(chan_q)
    wg.Wait()
}

func transmitter(CQ *cq.ChanQueue) {
    var cnt uint32
    cnt = 0
    defer wg.Done()
    // Open interface
    handle, err = pcap.OpenLive(device, 1500, promiscuous, timeout)
    if err != nil { log.Fatal(err) }
    defer handle.Close()
    loop:
    for {
        // check queueu, if queue isn't empty, than sneding packet to interface.
        if val, check := CQ.Pop(); check {
            handle.WritePacketData(val.([]byte))
            cnt++
        } else {
            select {
                // checking channel is close or not. If it closed, which mean that the all packets are sended.
                case <-end:
                    fmt.Println("transmit terminate")
                    break loop
                default:
                    continue
            }
        }
    }
    fmt.Println("transmitted : ", cnt)
}









