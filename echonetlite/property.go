package echonetlite

import "fmt"

// PropertyCode represents property code
type PropertyCode byte

// definition of property codes
const (
	// プロファイルオブジェクトスーパークラス
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

	// ノードプロファイルクラス
	// Class Group Code: 0x0E, Class Code: 0xF0
	NumOfInstances           PropertyCode = 0xD3 // 自ノードインスタンス数
	NumOfClasses             PropertyCode = 0xD4 // 自ノードクラス数
	InstanceListNotification PropertyCode = 0xD5 // インスタンスリスト通知
	InstanceListS            PropertyCode = 0xD6 // 自ノードインスタンスリストS
	ClassListS               PropertyCode = 0xD7 // 自ノードクラスリストS

	// 家庭用エアコンクラス
	MeasuredRoomTemperature    PropertyCode = 0xBB
	MeasuredOutdoorTemperature PropertyCode = 0xBE

	// 低圧スマート電力量メータクラス
	// Class Group Code: 0x02, Class Code: 0x88
	Coefficient                           PropertyCode = 0xD7 // 係数
	IntegralPowerConsumptionValidDigits   PropertyCode = 0xD7 // 積算電力量有効桁数
	IntegralPowerConsumption              PropertyCode = 0xE0 // 積算電力量計測値(正方向計測値)
	IntegralPowerConsumptionUnit          PropertyCode = 0xE1 // 積算電力量単位(正方向、逆方向計測値)
	IntegralPowerConsumptionHist1         PropertyCode = 0xE2 // 積算電力量計測値履歴１(正方向計測値)
	IntegralPowerConsumptionRev           PropertyCode = 0xE3 // 積算電力量計測値(逆方向計測値)
	IntegralPowerConsumptionRevHist1      PropertyCode = 0xE4 // 積算電力量計測値履歴１(逆方向計測値)
	IntegralPowerConsumptionHistCollDate1 PropertyCode = 0xE5 // 積算履歴収集日１
	InstantPower                          PropertyCode = 0xE7 // 瞬時電力計測値
	InstantCurrent                        PropertyCode = 0xE8 // 瞬時電流計測値
	PeriodicalIntegralPowerConsumption    PropertyCode = 0xEA // 定時積算電力量計測値(正方向計測値)
	PeriodicalIntegralPowerConsumptionRev PropertyCode = 0xEB // 定時積算電力量計測値(逆方向計測値)
	IntegralPowerConsumptionHist2         PropertyCode = 0xEC // 積算電力量計測値履歴２(正方向、逆方向計測値)
	IntegralPowerConsumptionHistCollDate2 PropertyCode = 0xED // 積算履歴収集日２

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

func (p Property) String() string {
	return fmt.Sprintf("EPC[%x] PDC[%d] EDT[%s]", p.Code, p.Len, p.Data)
}
