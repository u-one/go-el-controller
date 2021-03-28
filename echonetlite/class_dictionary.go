package echonetlite

import (
	"encoding/csv"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const (
	classInfoPath = "../../third-party/ECHONETLite-ObjectDatabase/data/csv/ja"
)

var (
	// classDictionary is a map with ClassGroup, Class as key and ClassInfo as value
	classDictionary ClassDictionary
)

// PrepareClassDictionary prepares information about Echonet Lite classes
func PrepareClassDictionary() error {
	classDictionary, err := load(classInfoPath)
	classDictionary.merge(loadNodeProfile(classInfoPath))
	classDictionary.merge(loadControllerProfile())
	return err
}

// GetClassDictionary returns ClassDictionary
func GetClassDictionary() ClassDictionary {
	return classDictionary
}

// ClassDictionary is Class keyed ClassInfo map
type ClassDictionary map[ClassGroupCode]map[ClassCode]ClassInfo

// ClassInfo is static information about Class
type ClassInfo struct {
	ClassGroup ClassGroupCode
	Class      ClassCode
	Properties PropertyDictionary
	Desc       string
}

type PropertyDictionary map[PropertyCode]PropertyInfo

// PropertyInfo is static information about property
type PropertyInfo struct {
	Code   PropertyCode
	Detail string
}

// NewClassDictionary returns ClassDictionary
func NewClassDictionary() ClassDictionary {
	return ClassDictionary{}
}

func (dict ClassDictionary) get(g ClassGroupCode, c ClassCode) (ClassInfo, bool) {
	if cm, ok := dict[g]; ok {
		if i, ok := cm[c]; ok {
			return i, true
		}
	}
	return ClassInfo{}, false
}

func (dict ClassDictionary) add(g ClassGroupCode, c ClassCode, info ClassInfo) {
	cm, ok := dict[g]
	if !ok {
		cm = map[ClassCode]ClassInfo{}
		dict[g] = cm
	}
	cm[c] = info
}

func (dict ClassDictionary) merge(other ClassDictionary) {
	for cg, cm := range other {
		for c, i := range cm {
			dict.add(cg, c, i)
		}
	}
}

// Get returns ClassInfo from Class key
func (dict ClassDictionary) Get(g ClassGroupCode, c ClassCode) ClassInfo {
	if i, ok := dict.get(g, c); ok {
		return i
	}
	return ClassInfo{
		ClassGroup: g,
		Class:      c,
		Properties: PropertyDictionary{},
		Desc:       "unknown",
	}
}

// load loads class information from files SonyCSL provides
// https://github.com/SonyCSL/ECHONETLite-ObjectDatabase
func load(basePath string) (ClassDictionary, error) {

	// There are files named in format 0xXXYY.csv (YY:class group code XX:class code)
	// DeviceList.csv
	// and DeviceObject.csv
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		logger.Println(err)
		return NewClassDictionary(), err
	}

	classMap := NewClassDictionary()

	for _, file := range files {
		codes := classCode(file) // 0xXXYY.csv
		if codes == nil {
			continue
		}
		logger.Println("Decoded class code", codes)

		logger.Println(file)
		logger.Println(basePath, file.Name())

		properties := loadClassInfo(basePath + "/" + file.Name())
		if properties != nil {
			clsInfo := ClassInfo{
				ClassGroup: ClassGroupCode(codes[0]),
				Class:      ClassCode(codes[1]),
				Properties: properties,
				Desc:       "",
			}
			classMap.add(clsInfo.ClassGroup, clsInfo.Class, clsInfo)
		}
	}

	return classMap, nil
}

func loadNodeProfile(basePath string) ClassDictionary {
	classMap := NewClassDictionary()

	properties := loadClassInfo(basePath + "/DeviceObject.csv")
	if properties != nil {
		properties[0xd3] = PropertyInfo{Code: 0xd3, Detail: "自ノードインスタンス数"}
		properties[0xd4] = PropertyInfo{Code: 0xd4, Detail: "自ノードクラス数"}
		properties[0xd5] = PropertyInfo{Code: 0xd5, Detail: "インスタンスリスト通知"}
		properties[0xd6] = PropertyInfo{Code: 0xd6, Detail: "自ノードインスタンスリストS"}
		properties[0xd7] = PropertyInfo{Code: 0xd7, Detail: "自ノードクラスリストS"}
		clsInfo := ClassInfo{
			ClassGroup: ClassGroupCode(0x0e),
			Class:      ClassCode(0xf0),
			Properties: properties,
			Desc:       "ノードプロファイル",
		}
		logger.Println(clsInfo)
		classMap.add(clsInfo.ClassGroup, clsInfo.Class, clsInfo)
	}
	return classMap
}

func loadControllerProfile() ClassDictionary {
	classMap := NewClassDictionary()

	clsInfo := ClassInfo{
		ClassGroup: ClassGroupCode(0x05),
		Class:      ClassCode(0xff),
		Properties: PropertyDictionary{},
		Desc:       "コントローラ",
	}
	logger.Println(clsInfo)
	classMap.add(clsInfo.ClassGroup, clsInfo.Class, clsInfo)
	return classMap
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

// loadPropertyInfo load PropertyInfo from file(0xXXYY.csv)
// which describes about property information for a Echonet Lite class
func loadClassInfo(filePath string) PropertyDictionary {

	properties := PropertyDictionary{}

	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		logger.Printf("failed to open file: %w", err)
		return properties
	}
	var epcBegan = false

	// csv format
	// Line 1-2: class info
	//   Line 1 Header: "Class name,Remarks,Group code,Class code,Whether or not detailed requirements are provided,,,,,,,"
	//   Line 2 Value : "Smart electric energy meter,,0x02,0x88,○,,,,,,,""
	// Line 3-5: Empty
	// Line 6-: property info
	//   Line 6 Header: "EPC,Property name,Contents of property,Value range(decimal notation),Unit,Data type,Data size,Access rule(Anno),Access rule(Set),Access rule(Get),Announcement at status change,Remark"
	//   Line 7 Value: "0x80,Operation status,This property indicates the ON/OFF status.,"ON=0x30, OFF=0x31",.,unsigned char,1,-,optional,mandatory,mandatory,"
	//   ...

	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Println(err)
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
		properties[PropertyCode(d[0])] = p
	}

	return properties
}
