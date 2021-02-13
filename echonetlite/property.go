package echonetlite

// PropertyCode represents property code
type PropertyCode byte

// definition of property codes
const (
	OperationStatus                PropertyCode = 0x80 // 動作状態
	InstallationLocation           PropertyCode = 0x81 // 設置場所
	SpecVersion                    PropertyCode = 0x82 // 規格Version情報
	ID                             PropertyCode = 0x83 // 識別番号
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
