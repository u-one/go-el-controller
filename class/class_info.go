package class

import (
	"encoding/csv"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var logger *log.Logger

func init() {
	//logger = log.New(os.Stdout, "[ClassInfo]", log.LstdFlags)
	logger = log.New(ioutil.Discard, "[ClassInfo]", log.LstdFlags)
}

type ClassInfoMap map[ClassCode]*ClassInfo

func (cim ClassInfoMap) Get(code ClassCode) *ClassInfo {
	if cim[code] != nil {
		return cim[code]
	}
	return &ClassInfo{
		ClassGroupCode: code.ClassGroupCode,
		ClassCode:      code.ClassCode,
		Properties:     make(map[PropertyCode]*PropertyInfo),
		Desc:           "unknown",
	}
}

// ClassCode is Echonet-Lite Class information
type ClassCode struct {
	ClassGroupCode byte
	ClassCode      byte // 0xF0
}

// NewClassCode returns new instance of ClassCode
func NewClassCode(classGroupCode, classCode byte) ClassCode {
	logger.Println("NewClassCode ClassGroup code:", classGroupCode, " Class code:", classCode)
	return ClassCode{
		ClassGroupCode: classGroupCode,
		ClassCode:      classCode,
	}
}

type ClassInfo struct {
	ClassGroupCode byte
	ClassCode      byte
	Properties     map[PropertyCode]*PropertyInfo
	Desc           string
}

func NewClassInfo() *ClassInfo {
	c := ClassInfo{}
	return &c
}

type PropertyCode byte

type PropertyInfo struct {
	Code   PropertyCode
	Detail string
}

// ex.
// SEOJ 0x0ef001
// class group code: 0e
// class code: f0
// instance: 01
// EPC 0x80
// property: 80
func Load() (ClassInfoMap, error) {
	path := "../../SonyCSL/ECHONETLite-ObjectDatabase/data/csv/ja"
	files, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}

	classMap := make(ClassInfoMap)

	for _, file := range files {
		codes := classCode(file)
		if codes == nil {
			continue
		}
		logger.Println("Decoded class code", codes)

		logger.Println(file)
		logger.Println(path, file.Name())

		properties := loadFromFile(path + "/" + file.Name())
		if properties != nil {
			cls := ClassInfo{
				ClassGroupCode: codes[0],
				ClassCode:      codes[1],
				Properties:     properties,
				Desc:           "",
			}
			classMap[NewClassCode(cls.ClassGroupCode, cls.ClassCode)] = &cls
		}
	}

	properties := loadFromFile(path + "/DeviceObject.csv")
	if properties != nil {
		properties[0xd3] = &PropertyInfo{Code: 0xd3, Detail: "自ノードインスタンス数"}
		properties[0xd4] = &PropertyInfo{Code: 0xd4, Detail: "自ノードクラス数"}
		properties[0xd5] = &PropertyInfo{Code: 0xd5, Detail: "インスタンスリスト通知"}
		properties[0xd6] = &PropertyInfo{Code: 0xd6, Detail: "自ノードインスタンスリストS"}
		properties[0xd7] = &PropertyInfo{Code: 0xd7, Detail: "自ノードクラスリストS"}
		cls := ClassInfo{
			ClassGroupCode: 0x0e,
			ClassCode:      0xf0,
			Properties:     properties,
			Desc:           "ノードプロファイル",
		}
		logger.Println(cls)
		classMap[NewClassCode(cls.ClassGroupCode, cls.ClassCode)] = &cls
	}

	cls := ClassInfo{
		ClassGroupCode: 0x05,
		ClassCode:      0xff,
		Properties:     make(map[PropertyCode]*PropertyInfo, 0),
		Desc:           "コントローラ",
	}
	logger.Println(cls)
	classMap[NewClassCode(cls.ClassGroupCode, cls.ClassCode)] = &cls

	return classMap, nil
}

func classCode(file os.FileInfo) []byte {
	name := strings.Split(file.Name(), ".")[0]

	if !strings.HasPrefix(name, "0x") {
		logger.Println("Not property file: ", name)
		return nil
	}

	decodedClassCodes, err := hex.DecodeString(strings.TrimPrefix(name, "0x"))
	if err != nil {
		logger.Println(err)
		return nil
	}
	logger.Println(decodedClassCodes)
	return decodedClassCodes
}

func loadFromFile(filePath string) map[PropertyCode]*PropertyInfo {

	properties := make(map[PropertyCode]*PropertyInfo, 0)

	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		logger.Fatal(err)
		return nil
	}
	var epcBegan = false

	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Fatal(err)
			continue
		}
		if record[0] == "EPC" {
			epcBegan = true
			continue
		}
		if !epcBegan {
			continue
		}
		logger.Println(record[0])
		if !strings.HasPrefix(record[0], "0x") {
			continue
		}
		if len(record[0]) == 0 {
			continue
		}
		d, err := hex.DecodeString(strings.TrimPrefix(record[0], "0x"))
		if err != nil {
			logger.Println("failed to decode:", record[0])
			continue
		}

		p := PropertyInfo{
			Code:   PropertyCode(d[0]),
			Detail: record[1],
		}
		properties[PropertyCode(d[0])] = &p
	}

	return properties
}
