package echonetlite

// PropertyCode represents property code
type PropertyCode byte

// definition of property codes
const (
	InstallationLocation           PropertyCode = 0x81 // 設置場所
	MomentaryPowerConsumption      PropertyCode = 0x84 // 瞬間消費電力計測値
	IntegratingPowerConsumption    PropertyCode = 0x85 // 積算消費電力計測値
	ManufacturerErrorCode          PropertyCode = 0x86 // メーカ異常コード
	ElectricityCurrentLimit        PropertyCode = 0x87 // 電流制限設定
	AbnormalState                  PropertyCode = 0x88 // 異常発生状態
	AbnormalDetail                 PropertyCode = 0x89 // 異常発生内容
	ManufacturerCode               PropertyCode = 0x8A
	OfficePlaceCode                PropertyCode = 0x8B
	ProductCode                    PropertyCode = 0x8C
	ManufacturingNumber            PropertyCode = 0x8D
	ManufacturingDate              PropertyCode = 0x8E
	PowerReductionState            PropertyCode = 0x8F
	RemoteControlState             PropertyCode = 0x93
	CurrentTime                    PropertyCode = 0x97
	CurrentDate                    PropertyCode = 0x98
	PowerConsumptionLimit          PropertyCode = 0x99
	IntegratingOperatingTime       PropertyCode = 0x9A
	SetMPropertyMap                PropertyCode = 0x9B // SetMプロパティマップ
	GetMPropertyMap                PropertyCode = 0x9C // GetMプロパティマップ
	StageChangeAnnouncePropertyMap PropertyCode = 0x9D // 状変アナウンスプロパティマップ
	SetPropertyMap                 PropertyCode = 0x9E // Setプロパティマップ
	GetPropertyMap                 PropertyCode = 0x9F // Getプロパティマップ

	MeasuredRoomTemperature    PropertyCode = 0xBB
	MeasuredOutdoorTemperature PropertyCode = 0xBE

	// ノードプロファイルクラス
	// Class Group Code: 0x0E, Class Code: 0xF0
	OperationStatus          PropertyCode = 0x80 // 動作状態
	SpecVersion              PropertyCode = 0x82 // 規格Version情報
	ID                       PropertyCode = 0x83 // 識別番号
	NumOfInstances           PropertyCode = 0xD3 // 自ノードインスタンス数
	NumOfClasses             PropertyCode = 0xD4 // 自ノードクラス数
	InstanceListNotification PropertyCode = 0xD5 // インスタンスリスト通知
	InstanceListS            PropertyCode = 0xD6 // 自ノードインスタンスリストS
	ClassListS               PropertyCode = 0xD7 // 自ノードクラスリストS

	// 低圧スマート電力量メータクラス
	// Class Group Code: 0x02, Class Code: 0x88

)

// Property represents Echonet-Lite property
type Property struct {
	Code byte
	Len  int
	Data Data
}

// Serialize returns serialized data
func (p Property) Serialize() Data {
	d := []byte{}
	d = append(d, p.Code)
	d = append(d, byte(p.Len))
	d = append(d, p.Data...)
	return d
}
