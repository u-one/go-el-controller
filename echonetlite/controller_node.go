package echonetlite

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/u-one/go-el-controller/transport"
)

var clogger *log.Logger

func init() {
	clogger = log.New(os.Stdout, "[Controller]", log.LstdFlags)
}

var (
	tempMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "home",
			Subsystem: "aircon",
			Name:      "temperature",
			Help:      "aircon temp",
		},
		[]string{
			"ip", "type", "location",
		},
	)
)

func init() {
	prometheus.MustRegister(tempMetrics)

}

const (
	// MulticastIP is Echonet-Lite multicast address
	MulticastIP = "224.0.23.0"
	// Port is Echonet-Lite receive port
	Port = ":3610"
)

// ControllerNode is ECHONETLite controller
type ControllerNode struct {
	MulticastReceiver transport.MulticastReceiver
	UnicastReceiver   transport.UnicastReceiver
	MulticastSender   transport.MulticastSender
	tid               uint16
	nodeList          NodeList
}

// NewControllerNode returns ControllerNode
func NewControllerNode() (*ControllerNode, error) {
	ms, err := transport.NewUDPMulticastSender(MulticastIP, Port)
	if err != nil {
		log.Println(err)
		return &ControllerNode{}, err
	}
	return &ControllerNode{
		MulticastReceiver: &transport.UDPMulticastReceiver{},
		MulticastSender:   ms,
		UnicastReceiver:   &transport.UDPUnicastReceiver{},
	}, nil
}

// Close closes all resources open
func (elc ControllerNode) Close() {
	elc.MulticastSender.Close()
}

// NodeList is list of node profile objects
type NodeList map[string]Node

// Add adds Node
func (nlist *NodeList) Add(addr string, obj Object) {
	if _, ok := (*nlist)[addr]; !ok {
		(*nlist)[addr] = Node{}
	}
}

// Node represents a node profile object
type Node struct {
	Devices []Object
}

// Start starts controller
func (elc ControllerNode) Start(ctx context.Context) {
	elc.tid = 0
	elc.nodeList = make(NodeList)

	sch := elc.UnicastReceiver.Start(ctx, Port)
	go elc.handleUnicastResult(ctx, sch)

	mch := elc.MulticastReceiver.Start(ctx, MulticastIP, Port)
	go elc.handleMulticastResult(ctx, mch)

	elc.startSequence(ctx)
}

func (elc ControllerNode) handleMulticastResult(ctx context.Context, results <-chan transport.ReceiveResult) {
	for {
		select {
		case <-ctx.Done():
			clogger.Println("readMulticast handler ctx.Done")
			return
		case result := <-results:
			if result.Err != nil {
				clogger.Printf("[Error] failed to receive [%s]\n", result.Err)
				break
			}
			clogger.Printf("<<<<<<<< [%v] MULTI CAST RECEIVE: ", result.Address)
			err := elc.onReceive(ctx, result)
			if err != nil {
				clogger.Printf("[Error] %s", err)
				break
			}
		}
	}
}

func (elc ControllerNode) handleUnicastResult(ctx context.Context, results <-chan transport.ReceiveResult) {
	for {
		select {
		case <-ctx.Done():
			clogger.Println("readUnicast handler ctx.Done")
			return
		case result := <-results:
			if result.Err != nil {
				clogger.Printf("[Error] failed to receive [%s]\n", result.Err)
				break
			}
			clogger.Printf("<<<<<<<< [%v] UNI CAST RECEIVE: ", result.Address)
			err := elc.onReceive(ctx, result)
			if err != nil {
				clogger.Printf("[Error] %s", err)
				break
			}
		}
	}
}

func (elc ControllerNode) onReceive(ctx context.Context, recv transport.ReceiveResult) error {
	frame, err := ParseFrame(recv.Data)
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}
	clogger.Printf("[%v] %s\n", recv.Address, frame)

	var targetClass Class
	if frame.ESV.isResponseOrNotification() {
		targetClass = frame.SrcClass()
	} else {
		targetClass = frame.DstClass()
	}
	obj, err := parseProperties(targetClass, frame.Properties)
	if err != nil {
		return fmt.Errorf("ParseProperties failed: %w", err)
	}

	switch frame.ESV {
	// 要求
	case SetI, // プロパティ値書き込み(応答不要)
		SetC,   // プロパティ値書き込み(応答要)
		Get,    // プロパティ値読み出し
		InfReq, // プロパティ値通知要求
		SetGet: // プロパティ値書き込み・読み出し要求
	// 応答・通知
	case SetRes: // プロパティ値書き込み
	case GetRes: // プロパティ値読み出し応答
		switch obj.(type) {
		case AirconObject:
			o := obj.(AirconObject)
			lc := o.InstallLocation.Code
			ln := o.InstallLocation.Number
			loc := lc.String()
			if ln != 0 {
				loc = fmt.Sprintf("%s%d", lc, ln)
			}

			tempMetrics.With(prometheus.Labels{"ip": recv.Address, "location": loc, "type": "room"}).Set(o.InternalTemp)
			tempMetrics.With(prometheus.Labels{"ip": recv.Address, "location": loc, "type": "outside"}).Set(o.OuterTemp)

		}
	case Inf: // プロパティ値通知
		elc.nodeList.Add(recv.Address, frame.SEOJ)
		//[Controller]2019/09/27 01:52:59 [192.168.1.15] 108100010ef00105ff017301d50401013001 EHD[1081] TID[0001] SEOJ[0ef001](ノードプロファイル) DEOJ[05ff01](コントローラ) ESV[INF] OPC[01] EPC0[d5](インスタンスリスト通知) PDC0[4] EDT0[01013001]
		//[Controller]2019/09/27 01:52:59 [192.168.1.10] 108100010ef00105ff017301d50401013001 EHD[1081] TID[0001] SEOJ[0ef001](ノードプロファイル) DEOJ[05ff01](コントローラ) ESV[INF] OPC[01] EPC0[d5](インスタンスリスト通知) PDC0[4] EDT0[01013001]
	case InfC: //
	case InfCRes: //
	case SetGetRes: //
	case SetISNA: //
	case SetCSNA: //
	case GetSNA: //
	//[Controller]2021/02/23 16:58:06 [192.168.50.102] 108100020ef00105ff0152088001308204010c0100d303000001d4020002d500d60401013001d7030101309f0e0d808283898a9d9e9fbfd3d4d6d7 EHD[1081] TID[0002] SEOJ[{0ef001}](unknown) DEOJ[{05ff01}](unknown) ESV[Get_SNA] OPC[8] EPC0[80]() PDC0[1] EDT0[30] EPC1[82]() PDC1[4] EDT1[010c0100] EPC2[d3]() PDC2[3] EDT2[000001] EPC3[d4]() PDC3[2] EDT3[0002] EPC4[d5]() PDC4[0] EDT4[] EPC5[d6]() PDC5[4] EDT5[01013001] EPC6[d7]() PDC6[3] EDT6[010130] EPC7[9f]() PDC7[14] EDT7[0d808283898a9d9e9fbfd3d4d6d7]
	case InfSNA: //
	case SetGetSNA: //
	}
	return nil
}

func (elc *ControllerNode) sendFrame(f *Frame) {
	clogger.Printf(">>>>>>>> SEND : %s\n", f)
	elc.MulticastSender.Send([]byte(f.Serialize()))
	elc.tid++
}

func (elc *ControllerNode) startSequence(ctx context.Context) {
	clogger.Println("Start Sequnce Begin")

	f := CreateInfFrame(elc.tid)
	elc.sendFrame(f)

	// ver.1.0
	f = CreateInfReqFrame(elc.tid)
	elc.sendFrame(f)

	// ver.1.1
	f = CreateGetFrame(elc.tid)
	elc.sendFrame(f)

	time.Sleep(time.Second * 3)
	clogger.Println("Start Sequnce End")
}

// RequestAirConState sends request to get air conditioner states
func (elc *ControllerNode) RequestAirConState() {
	f := CreateAirconGetFrame(elc.tid)
	elc.sendFrame(f)
}
