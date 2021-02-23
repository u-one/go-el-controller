package echonetlite

import "encoding/hex"

// ESVType represnts type of ESV
type ESVType byte

// ESVTypes
const (
	// 要求
	SetI   ESVType = 0x60 // SetI プロパティ値書き込み要求(応答不要)
	SetC   ESVType = 0x61 // SetC プロパティ値書き込み要求(応答要)
	Get    ESVType = 0x62 // Get プロパティ値読み出し要求
	InfReq ESVType = 0x63 // INF_REQ プロパティ値通知要求
	SetGet ESVType = 0x6E // SetGet プロパティ値書き込み・読み出し要求

	// 応答・通知
	SetRes    ESVType = 0x71 // Set_Res プロパティ値書き込み応答
	GetRes    ESVType = 0x72 // Get_Res プロパティ値読み出し応答
	Inf       ESVType = 0x73 // INF プロパティ値通知
	InfC      ESVType = 0x74 // INFC
	InfCRes   ESVType = 0x7A // INFC_Res
	SetGetRes ESVType = 0x7E // SetGet_Res

	// 不可応答
	SetISNA   ESVType = 0x50 // SetI_SNA
	SetCSNA   ESVType = 0x51 // SetC_SNA
	GetSNA    ESVType = 0x52 // Get_SNA
	InfSNA    ESVType = 0x53 // INF_SNA
	SetGetSNA ESVType = 0x5E // SetGet_SNA
)

func (t ESVType) String() string {
	switch t {
	case SetI:
		return "SetI"
	case SetC:
		return "SetC"
	case Get:
		return "Get"
	case InfReq:
		return "INF_REQ"
	case SetGet:
		return "SetGet"
	case SetRes:
		return "Set_Res"
	case GetRes:
		return "Get_Res"
	case Inf:
		return "INF"
	case InfC:
		return "INFC"
	case InfCRes:
		return "INFC_Res"
	case SetGetRes:
		return "SetGet_Res"
	case SetISNA:
		return "SetI_SNA"
	case SetCSNA:
		return "SetC_SNA"
	case GetSNA:
		return "Get_SNA"
	case InfSNA:
		return "INF_SNA"
	case SetGetSNA:
		return "SetGet_SNA"
	default:
		return "UNKNOWN" + hex.EncodeToString([]byte{byte(t)})
	}
}

func (t ESVType) isRequest() bool {
	switch t {
	case SetI,
		SetC,
		Get,
		InfReq,
		SetGet:
		return true
	}
	return false
}

func (t ESVType) isResponseOrNotification() bool {
	switch t {
	case SetRes,
		GetRes,
		Inf,
		InfC,
		InfCRes,
		SetGetRes,
		SetISNA,
		SetCSNA,
		GetSNA,
		InfSNA,
		SetGetSNA:
		return true
	}
	return false
}
