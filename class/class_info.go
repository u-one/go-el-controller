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

type ClassInfoMap map[ClassCode]*ClassInfo

// ClassCode is Echonet-Lite Class information
type ClassCode struct {
	ClassGroupCode byte
	ClassCode      byte // 0xF0
}

// NewClassCode returns new instance of ClassCode
func NewClassCode(classGroupCode, classCode byte) ClassCode {
	log.Println("NewClassCode ClassGroup code:", classGroupCode, " Class code:", classCode)
	return ClassCode{
		ClassGroupCode: classGroupCode,
		ClassCode:      classCode,
	}
}

type ClassInfo struct {
	ClassGroupCode byte
	ClassCode      byte
	Properties     map[PropertyCode]*PropertyInfo
	Description    string
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
func Load() ClassInfoMap {
	path := "../../SonyCSL/ECHONETLite-ObjectDatabase/data/csv/ja"
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	classMap := make(ClassInfoMap)

	for _, file := range files {
		codes := classCode(file)
		if codes == nil {
			continue
		}
		log.Println("Decoded class code", codes)

		log.Println(file)
		log.Println(path, file.Name())

		properties := loadFromFile(path + "/" + file.Name())
		if properties != nil {
			cls := ClassInfo{
				ClassGroupCode: codes[0],
				ClassCode:      codes[1],
				Properties:     properties,
			}
			classMap[NewClassCode(cls.ClassGroupCode, cls.ClassCode)] = &cls
		}
	}

	properties := loadFromFile(path + "/DeviceObject.csv")
	if properties != nil {
		cls := ClassInfo{
			ClassGroupCode: 0x0e,
			ClassCode:      0xf0,
			Properties:     properties,
			Description:    "ノードプロファイル",
		}
		log.Println(cls)
		classMap[NewClassCode(cls.ClassGroupCode, cls.ClassCode)] = &cls
	}

	return classMap
}

func classCode(file os.FileInfo) []byte {
	name := strings.Split(file.Name(), ".")[0]

	if !strings.HasPrefix(name, "0x") {
		log.Println("Not property file: ", name)
		return nil
	}

	decodedClassCodes, err := hex.DecodeString(strings.TrimPrefix(name, "0x"))
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Println(decodedClassCodes)
	return decodedClassCodes
}

func loadFromFile(filePath string) map[PropertyCode]*PropertyInfo {

	properties := make(map[PropertyCode]*PropertyInfo, 0)

	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
			continue
		}
		if record[0] == "EPC" {
			epcBegan = true
			continue
		}
		if !epcBegan {
			continue
		}
		log.Println(record[0])
		if !strings.HasPrefix(record[0], "0x") {
			continue
		}
		if len(record[0]) == 0 {
			continue
		}
		d, err := hex.DecodeString(strings.TrimPrefix(record[0], "0x"))
		if err != nil {
			log.Println("failed to decode:", record[0])
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
