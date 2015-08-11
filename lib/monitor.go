package monitor

import (
    "fmt"
    "net"
    "syscall"

    "encoding/binary"
)

func RegisterIP(ip *net.IPNet) {

}

func MonitorIP() {
    // Create a Unix socket to listen for NetLink changes
    fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_ROUTE)
    if err != nil {
        fmt.Println("Error creating unix socket:", err)
        return
    }
    defer syscall.Close(fd)

    // Netlink options
    sn := syscall.SockaddrNetlink{
        Pid:    0,                                                  // Pid 0 == KERNEL messages
        Groups: syscall.RTNLGRP_LINK | syscall.RTNLGRP_IPV4_IFADDR, // Broadcast groups
    }

    // Bind to the Socket
    err = syscall.Bind(fd, &sn)
    if err != nil {
        fmt.Println("Binding fd to socket failed")
    }

    // Infinite loop to keep listening for changes
    buff := make([]byte, syscall.Getpagesize())
    for {
        n, _, err := syscall.Recvfrom(fd, buff, 0x00)

        if err != nil {
            fmt.Println("Error when receiving from the socket: ", err)
        }

        // Incorrect header length, ignore
        if n < syscall.NLMSG_HDRLEN {
            continue
        }

        msgs, err := syscall.ParseNetlinkMessage(buff[:n])
        if err != nil {
            fmt.Println("Parsing netlink message failed")
        }

        // Loop over parsed Netlink messages
        for _, msg := range msgs {

            if msg.Header.Type == syscall.NLMSG_ERROR {
                fmt.Println("Netlink message error!")
                continue
            }

            // Ignore multipart messages
            if msg.Header.Type == syscall.NLMSG_DONE {
                continue
            }

            // Ignore neighbor messages
            if msg.Header.Type == syscall.RTM_DELNEIGH || msg.Header.Type == syscall.RTM_NEWNEIGH {
                continue
            }

            fmt.Printf("RECEIVED : %+v \n", msg)

            if msg.Header.Type == syscall.RTM_NEWLINK {
                ifMsg := syscall.IfInfomsg{
                    Family:     msg.Data[0], // AF_UNSPEC (always)
                    X__ifi_pad: msg.Data[1], // Reserved
                    Type:       binary.LittleEndian.Uint16(msg.Data[2:4]),
                    Index:      int32(binary.LittleEndian.Uint32(msg.Data[4:8])),
                    Flags:      binary.LittleEndian.Uint32(msg.Data[8:12]),
                    Change:     binary.LittleEndian.Uint32(msg.Data[12:16]), // Reserved for future use
                }

                // Check if the interface is "administratively" up
                if ifMsg.Flags&syscall.IFF_UP == 0 {
                    continue
                }

                // Check if the interface is operationally up
                if ifMsg.Flags&syscall.IFF_RUNNING == 0 {
                    continue
                }

                //nlAttrs, _ := syscall.ParseNetlinkRouteAttr(&msg)
                iface, _ := net.InterfaceByIndex(int(ifMsg.Index))
                addrs, _ := iface.Addrs()

                //fmt.Printf("RTM_NEWLINK: %+v \n", ifMsg)
                //fmt.Printf("RTM_NEWLINK ATTRIBUTES: %+v \n", nlAttrs)
                //fmt.Printf("RTM_NEWLINK INTERFACE: %+v \n", iface)
                fmt.Printf("RTM_NEWLINK INTERFACE ADDRS: %+v \n", addrs)

                if len(addrs) > 0 {
                    fmt.Printf("%s is now available at %s \n", iface.Name, addrs[0].(*net.IPNet).IP)
                } else {
                    fmt.Printf("No address associated yet with %s \n", iface.Name)
                }
            } else {
                continue
            }

        }

    }
}
