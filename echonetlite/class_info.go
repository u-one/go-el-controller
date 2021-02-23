package echonetlite

import (
	"encoding/csv"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	// ClassInfoDB is a map with Class as key and ClassInfo as value
	// TODO: refactor
	ClassInfoDB ClassInfoMap
)

// ClassInfoMap is Class keyed ClassInfo map
type ClassInfoMap map[Class]*ClassInfo

// Get returns ClassInfo from Class key
func (cim ClassInfoMap) Get(c Class) *ClassInfo {
	if cim[c] != nil {
		return cim[c]
	}
	return &ClassInfo{
		Class:      Class{c.ClassGroupCode, c.ClassCode},
		Properties: make(map[PropertyCode]*PropertyInfo),
		Desc:       "unknown",
	}
}

// Class is Echonet-Lite Class information
type Class struct {
	ClassGroupCode byte
	ClassCode      byte // 0xF0
}

// NewClass returns new instance of Class
func NewClass(o Object) Class {
	log.Println("[ClassInfo]NewClass ClassGroup code:", o.classGroupCode(), " Class code:", o.classCode())
	return Class{
		ClassGroupCode: byte(o.classGroupCode()),
		ClassCode:      byte(o.classCode()),
	}
}

// NewClassWithCode returns new instance of Class
func NewClassWithCode(cgc ClassGroupCode, cc ClassCode) Class {
	log.Println("[ClassInfo]NewClass ClassGroup code:", cgc, " Class code:", cc)
	return Class{
		ClassGroupCode: byte(cgc),
		ClassCode:      byte(cc),
	}
}

// ClassInfo is static information about Class
type ClassInfo struct {
	Class      Class
	Properties map[PropertyCode]*PropertyInfo
	Desc       string
}

// NewClassInfo creates ClassInfo instance
func NewClassInfo() *ClassInfo {
	c := ClassInfo{}
	return &c
}

// PropertyInfo is static information about property
type PropertyInfo struct {
	Code   PropertyCode
	Detail string
}

// Load loads class information from files and creates ClassInfoMap
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
		logger.Println(err)
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
			clsInfo := ClassInfo{
				Class:      Class{codes[0], codes[1]},
				Properties: properties,
				Desc:       "",
			}
			classMap[clsInfo.Class] = &clsInfo
		}
	}

	properties := loadFromFile(path + "/DeviceObject.csv")
	if properties != nil {
		properties[0xd3] = &PropertyInfo{Code: 0xd3, Detail: "自ノードインスタンス数"}
		properties[0xd4] = &PropertyInfo{Code: 0xd4, Detail: "自ノードクラス数"}
		properties[0xd5] = &PropertyInfo{Code: 0xd5, Detail: "インスタンスリスト通知"}
		properties[0xd6] = &PropertyInfo{Code: 0xd6, Detail: "自ノードインスタンスリストS"}
		properties[0xd7] = &PropertyInfo{Code: 0xd7, Detail: "自ノードクラスリストS"}
		clsInfo := ClassInfo{
			Class:      Class{0x0e, 0xf0},
			Properties: properties,
			Desc:       "ノードプロファイル",
		}
		logger.Println(clsInfo)
		classMap[clsInfo.Class] = &clsInfo
	}

	clsInfo := ClassInfo{
		Class:      Class{0x05, 0xff},
		Properties: make(map[PropertyCode]*PropertyInfo, 0),
		Desc:       "コントローラ",
	}
	logger.Println(clsInfo)
	classMap[clsInfo.Class] = &clsInfo

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
