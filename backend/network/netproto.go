package network

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"wakelan/backend/comm"
	"wakelan/backend/db"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type IpInfo struct {
	IP    net.IP
	Mac   net.HardwareAddr
	MANUF string
}

type ArpRetFun func(info IpInfo)
type PingRetFun func(ip string, mac string)

type NetProto struct {
	iface     *net.Interface
	handle    *pcap.Handle
	ipinfos   map[string]IpInfo
	lock      sync.Mutex
	cancelFun context.CancelFunc
	ctx       context.Context
	openLock  sync.Mutex
	pingFuns  map[string]PingRetFun
	arpFun    ArpRetFun
}

func (n *NetProto) Init() error {
	netCard := db.GetNetworkCard()
	if len(netCard) != 0 {
		data := map[string]interface{}{}
		err := json.Unmarshal([]byte(netCard), &data)
		if err == nil {
			name := data["name"].(string)
			if len(name) != 0 {
				iface, err := n.GetInterfaceByName(name)
				if err == nil {
					n.Close()
					n.Open(iface, true)
				}
			}
		}
	}

	n.pingFuns = make(map[string]PingRetFun)

	return nil
}

func (n *NetProto) GetLocalInfo() *net.Interface {
	return n.iface
}

func (n *NetProto) AddPingRetFun(flag string, fun PingRetFun) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.pingFuns[flag] = fun
}

func (n *NetProto) DelPingRetFun(flag string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	delete(n.pingFuns, flag)
}

func (n *NetProto) SetArpRetFun(flag string, fun ArpRetFun) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.arpFun = fun
}

func (n *NetProto) GetInterfaces() ([]pcap.Interface, error) {
	ifaces, err := pcap.FindAllDevs()
	if err != nil {
		return []pcap.Interface{}, err
	}

	tifaces := []pcap.Interface{}

	for _, v := range ifaces {
		if len(v.Addresses) == 0 {
			continue
		}

		if len(v.Description) == 0 {
			v.Description = v.Name
		}

		tifaces = append(tifaces, v)
	}

	return tifaces, nil
}

func (n *NetProto) GetInterfaceByName(name string) (pcap.Interface, error) {
	vs, err := n.GetInterfaces()
	if err != nil {
		return pcap.Interface{}, err
	}

	for _, v := range vs {
		if strings.EqualFold(v.Name, name) {
			return v, nil
		}
	}

	return pcap.Interface{}, errors.New("no find")
}

func (n *NetProto) makePingPkg(srcMac net.HardwareAddr, srcIP, dstIP net.IP) ([]byte, error) {
	eth := layers.Ethernet{
		SrcMAC:       srcMac,                                               // 源 MAC 地址
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // 目标 MAC 地址
		EthernetType: layers.EthernetTypeIPv4,
	}

	ipLayer := layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Version:  4,
		TTL:      64,
		Id:       uint16(os.Getgid()),
		Protocol: layers.IPProtocolICMPv4,
		Flags:    layers.IPv4DontFragment,
	}

	icmpLayer := layers.ICMPv4{
		TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoRequest, 0),
		Id:       uint16(os.Getpid()),
		Seq:      1,
	}

	data := []byte("abcdefghijklmnopqrstuvwabcdefghi")

	opt := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}

	buf := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buf, opt, &eth, &ipLayer, &icmpLayer, gopacket.Payload(data))
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func (n *NetProto) makeNetProtoPkg(srcMac net.HardwareAddr, srcIP, dstIP net.IP) ([]byte, error) {
	eth := layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}

	netproto := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     uint8(len(srcMac)),
		ProtAddressSize:   uint8(len(srcIP.To4())),
		Operation:         layers.ARPRequest,
		SourceProtAddress: srcIP.To4(),
		SourceHwAddress:   srcMac,
		DstHwAddress:      net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		DstProtAddress:    dstIP.To4(),
	}

	opt := gopacket.SerializeOptions{}

	buf := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buf, opt, &eth, &netproto)
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func (n *NetProto) IsOpen() bool {
	n.openLock.Lock()
	defer n.openLock.Unlock()

	return n.handle != nil
}

func (n *NetProto) Open(iface pcap.Interface, promisc bool) error {
	n.Close()

	n.openLock.Lock()
	defer n.openLock.Unlock()

	infos, err := net.Interfaces()
	if err != nil {
		return err
	}

	handle, err := pcap.OpenLive(iface.Name, 65536, promisc, pcap.BlockForever)
	if err != nil {
		return err
	}

	ctx, cancelFun := context.WithCancel(context.Background())

	n.handle = handle
	n.cancelFun = cancelFun
	n.ctx = ctx
	n.ipinfos = make(map[string]IpInfo)

	go func(h *pcap.Handle, ctx context.Context) {
		ps := gopacket.NewPacketSource(h, h.LinkType())

		for {
			select {
			case <-ctx.Done():
				return
			case p := <-ps.Packets():
				//处理netproto
				func() {
					pkg := p.Layer(layers.LayerTypeARP)
					if pkg == nil {
						return
					}

					netproto := pkg.(*layers.ARP)
					if netproto == nil {
						return
					}

					if netproto.Operation != layers.ARPReply {
						return
					}

					func() {
						info := IpInfo{}
						info.IP = netproto.SourceProtAddress
						info.Mac = netproto.SourceHwAddress
						info.MANUF = Search(info.Mac.String())

						if n.arpFun != nil {
							n.arpFun(info)
						}

						n.lock.Lock()
						defer n.lock.Unlock()
						n.ipinfos[info.Mac.String()] = info
					}()
				}()

				//处理icmp
				func() {
					pkg := p.Layer(layers.LayerTypeICMPv4)
					if pkg == nil {
						return
					}

					icmp := pkg.(*layers.ICMPv4)
					if icmp == nil {
						return
					}

					if icmp.TypeCode != layers.ICMPv4TypeEchoReply {
						return
					}

					tfuns := func() map[string]PingRetFun {
						n.lock.Lock()
						defer n.lock.Unlock()
						return n.pingFuns
					}()

					if len(tfuns) == 0 {
						return
					}

					pkg = p.Layer(layers.LayerTypeEthernet)
					eth := pkg.(*layers.Ethernet)
					pkg = p.Layer(layers.LayerTypeIPv4)
					ipLayer := pkg.(*layers.IPv4)

					go func(ip, mac string) {
						for _, fun := range tfuns {
							fun(ip, mac)
						}
					}(ipLayer.SrcIP.To4().String(), eth.SrcMAC.String())
				}()
			}
		}
	}(handle, ctx)

Open_Fin:
	for _, i := range infos {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			ip2, _, _ := net.ParseCIDR(addr.String())
			for _, ip := range iface.Addresses {
				if ip.IP.Equal(ip2) {
					n.iface = &i
					break Open_Fin
				}
			}
		}
	}

	if n.iface == nil {
		n.Close()
		return errors.New("net interface nil")
	}

	return nil
}

func (n *NetProto) Close() {
	n.openLock.Lock()
	defer n.openLock.Unlock()

	if n.cancelFun != nil {
		n.cancelFun()
		<-n.ctx.Done()
		n.cancelFun = nil

		if n.handle != nil {
			n.handle.Close()
			n.handle = nil
		}
	}

	n.lock.Lock()
	defer n.lock.Unlock()
	n.ipinfos = make(map[string]IpInfo)
}

func (n *NetProto) QueryIP(ip string) error {
	addrs, err := n.iface.Addrs()
	if err != nil {
		return err
	}

	tIP := net.ParseIP(ip)

	for _, addr := range addrs {
		srcIPNet := addr.(*net.IPNet)
		if !srcIPNet.Contains(tIP) {
			continue
		}

		pkg, err := n.makeNetProtoPkg(n.iface.HardwareAddr, srcIPNet.IP, tIP)
		if err != nil {
			return err
		}

		err = n.handle.WritePacketData(pkg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *NetProto) PingNet(ips []string) error {
	addrs, err := n.iface.Addrs()
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		srcIPNet := addr.(*net.IPNet)
		if srcIPNet.IP.To4() == nil {
			continue
		}

		for _, ip := range ips {
			ipObj := net.ParseIP(ip)
			pkg, err := n.makePingPkg(n.iface.HardwareAddr, srcIPNet.IP, ipObj)
			if err != nil {
				return err
			}

			err = n.handle.WritePacketData(pkg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *NetProto) QueryNet(millisecond int) error {
	addrs, err := n.iface.Addrs()
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		srcIPNet := addr.(*net.IPNet)
		if srcIPNet.IP.To4() == nil {
			continue
		}

		ips := comm.MakeIPs(*srcIPNet)
		for _, ip := range ips {
			pkg, err := n.makeNetProtoPkg(n.iface.HardwareAddr, srcIPNet.IP, ip)
			if err != nil {
				return err
			}

			err = n.handle.WritePacketData(pkg)
			if err != nil {
				return err
			}

			if millisecond != 0 {
				time.Sleep(time.Duration(millisecond) * time.Millisecond)
			}
		}
	}

	return nil
}

func (n *NetProto) GetResult() map[string]IpInfo {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.ipinfos
}

var netprotoOnce sync.Once
var netprotoObj *NetProto

func NetProtoObj() *NetProto {
	netprotoOnce.Do(func() {
		netprotoObj = &NetProto{}
	})

	return netprotoObj
}
