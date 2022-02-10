package main
import (
    "github.com/google/gopacket"
    "github.com/google/gopacket/pcap"
    cq "go_gen/queue"
    "flag"
    "log"
    "fmt"
    "time"
    "sync"
    "os"
    "os/signal"
    "syscall"
)

var (
    device       string
    err          error

    handle       *pcap.Handle
    buffer       gopacket.SerializeBuffer
    options      gopacket.SerializeOptions

    wg           sync.WaitGroup
    sigs         = make(chan os.Signal, 1)
    end          = make(chan struct{})
)

func init() {
    flag.StringVar(&device, "d", "", "Device name")
    flag.Parse()
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
}


func main() {
    chan_q := cq.NewChanQueue(1<<24)
    wg.Add(2)

    // Pakcet transciver
    go receiver(chan_q)
    go decoder(chan_q)
    wg.Wait()
}

func decoder(CQ *cq.ChanQueue) {
    defer wg.Done()
    cnt := 0
    Decode:
    for {
        if _, check := CQ.Pop_th(); check {
            if check {
                cnt++
                //fmt.Println(CQ.Len_th(), cnt)
            }
        } else {
            select {
                case <-end:
                    fmt.Println("decoders")
                    break Decode
                default:
            }
        }
    }
    fmt.Println("received : ", cnt)
}

func receiver(CQ *cq.ChanQueue) {
    defer wg.Done()
    /*
    // Open device
    handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
    if err != nil { log.Fatal(err) }
    defer handle.Close()
    */
    interHandle, err := pcap.NewInactiveHandle(device)
    if err != nil { log.Fatal(err) }

    err = interHandle.SetSnapLen(1500)
    if err != nil { log.Fatal(err) }
    err = interHandle.SetPromisc(true)
    if err != nil { log.Fatal(err) }
    err = interHandle.SetTimeout(time.Nanosecond)
    if err != nil { log.Fatal(err) }
    err = interHandle.SetBufferSize(1073741824) //1GB//67108864 64MB
    if err != nil { log.Fatal(err) }
    //err = interHandle.SetImmediateMode(true) //ImmediateMode will overrided Timeout function.
    //if err != nil { log.Fatal(err) }
    handle, err = interHandle.Activate()
    if err != nil { log.Fatal(err) }
    packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
//    packetSource.DecodeOptions.NoCopy = true
    tot_num := 0
    Receiver:
    for {
        select {
            case <- sigs:
                fmt.Println("sigs")
                close(end)
                break Receiver
            default:
                packet, err := packetSource.NextPacket()
                if err == nil {
                    tot_num++
                    CQ.Push_th(packet)
                    continue
                }
        }
    }
    fmt.Println(tot_num)
}
