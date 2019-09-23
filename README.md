
# golang EchonetLite Controller

and prometheus exporter

[![Build Status](https://travis-ci.org/u-one/go-el-controller.svg?branch=master)](https://travis-ci.org/u-one/go-el-controller)


### Environment

```
go get github.com/golang/mock/gomock
go install github.com/golang/mock/mockgen
```

clone object database file (though this repos has no go source code...)
```
go get github.com/SonyCSL/ECHONETLite-ObjectDatabase
```

### Build for Raspberry pi

```
env GOOS=linux GOARCH=arm GOARM=6 go build
```


### sample start sequence

```
[Frame]2019/09/25 01:46:36 {14 240 map[128:0xc000217f40 129:0xc000217f60 130:0xc000217f80 131:0xc000217fa0 132:0xc000217fc0 133:0xc000217fe0 134:0xc000248000 135:0xc000248020 136:0xc000248040 137:0xc000248060 138:0xc000248080 139:0xc0002480a0 140:0xc0002480c0 141:0xc0002480e0 142:0xc000248100 143:0xc000248120 147:0xc000248140 151:0xc000248160 152:0xc000248180 153:0xc0002481a0 154:0xc0002481c0 157:0xc0002481e0 158:0xc000248200 159:0xc000248220 211:0xc000248240 212:0xc000248260 213:0xc000248280 214:0xc0002482a0 215:0xc0002482c0] ノードプロファイル}
[Frame]2019/09/25 01:46:36 {5 255 map[] コントローラ}
[Controller]2019/09/25 01:46:36 startExporter:  :8083
2019/09/25 01:46:36 Start to listen multicast udp  224.0.23.0 :3610
[Frame]2019/09/25 01:46:36 frameReceived: {14 240}
2019/09/25 01:46:36 resolved: 224.0.23.0:3610
[Frame]2019/09/25 01:46:36 108100000ef0010ef0017301d5040105ff01 EHD[1081] TID[0000] SEOJ[0ef001](ノードプロファイル) DEOJ[0ef001](ノードプロファイル) ESV[INF] OPC[01] EPC0[d5](インスタンスリスト通知) PDC0[4] EDT0[0105ff01]
.[Controller]2019/09/25 01:46:36 >>>>>>>> sendFrame
[Frame]2019/09/25 01:46:36 108100000ef0010ef0017301d5040105ff01 EHD[1081] TID[0000] SEOJ[0ef001](ノードプロファイル) DEOJ[0ef001](ノードプロファイル) ESV[INF] OPC[01] EPC0[d5](インスタンスリスト通知) PDC0[4] EDT0[0105ff01]
2019/09/25 01:46:36 written: 18
[Frame]2019/09/25 01:46:36 frameReceived: {5 255}
[Frame]2019/09/25 01:46:36 1081000005ff010ef0016301d500 EHD[1081] TID[0000] SEOJ[05ff01](コントローラ) DEOJ[0ef001](ノードプロファイル) ESV[INF_REQ] OPC[01] EPC0[d5]() PDC0[0] EDT0[]
[Controller]2019/09/25 01:46:36 >>>>>>>> sendFrame
[Frame]2019/09/25 01:46:36 1081000005ff010ef0016301d500 EHD[1081] TID[0000] SEOJ[05ff01](コントローラ) DEOJ[0ef001](ノードプロファイル) ESV[INF_REQ] OPC[01] EPC0[d5]() PDC0[0] EDT0[]
2019/09/25 01:46:36 written: 14
[Frame]2019/09/25 01:46:36 frameReceived: {5 255}
[Frame]2019/09/25 01:46:36 1081000005ff010ef001620880008200d300d400d500d600d7009f00 EHD[1081] TID[0000] SEOJ[05ff01](コントローラ) DEOJ[0ef001](ノードプロファイル) ESV[Get] OPC[08] EPC0[80]() PDC0[0] EDT0[] EPC1[82]() PDC1[0] EDT1[] EPC2[d3]() PDC2[0] EDT2[] EPC3[d4]() PDC3[0] EDT3[] EPC4[d5]() PDC4[0] EDT4[] EPC5[d6]() PDC5[0] EDT5[] EPC6[d7]() PDC6[0] EDT6[] EPC7[9f]() PDC7[0] EDT7[]
[Controller]2019/09/25 01:46:36 >>>>>>>> sendFrame
[Frame]2019/09/25 01:46:36 1081000005ff010ef001620880008200d300d400d500d600d7009f00 EHD[1081] TID[0000] SEOJ[05ff01](コントローラ) DEOJ[0ef001](ノードプロファイル) ESV[Get] OPC[08] EPC0[80]() PDC0[0] EDT0[] EPC1[82]() PDC1[0] EDT1[] EPC2[d3]() PDC2[0] EDT2[] EPC3[d4]() PDC3[0] EDT3[] EPC4[d5]() PDC4[0] EDT4[] EPC5[d6]() PDC5[0] EDT5[] EPC6[d7]() PDC6[0] EDT6[] EPC7[9f]() PDC7[0] EDT7[]
2019/09/25 01:46:36 written: 28

.[Controller]2019/09/25 01:46:36 <<<<<<<< received
[Frame]2019/09/25 01:46:36 frameReceived: {14 240}
[Frame]2019/09/25 01:46:36 108100000ef00105ff017301d50401013001 EHD[1081] TID[0000] SEOJ[0ef001](ノードプロファイル) DEOJ[05ff01](コントローラ) ESV[INF] OPC[01] EPC0[d5](インスタンスリスト通知) PDC0[4] EDT0[01013001]
[Controller]2019/09/25 01:46:36 [192.168.1.10] 108100000ef00105ff017301d50401013001

.[Controller]2019/09/25 01:46:36 <<<<<<<< received
[Frame]2019/09/25 01:46:36 frameReceived: {14 240}
[Frame]2019/09/25 01:46:36 108100000ef00105ff0152088001308204010c0100d303000001d4020002d500d60401013001d7030101309f0e0d808283898a9d9e9fbfd3d4d6d7 EHD[1081] TID[0000] SEOJ[0ef001](ノードプロファイル) DEOJ[05ff01](コントローラ) ESV[Get_SNA] OPC[08] EPC0[80](動作状態) PDC0[1] EDT0[30] EPC1[82](規格Version情報) PDC1[4] EDT1[010c0100] EPC2[d3](自ノードインスタンス数) PDC2[3] EDT2[000001] EPC3[d4](自ノードクラス数) PDC3[2] EDT3[0002] EPC4[d5](インスタンスリスト通知) PDC4[0] EDT4[] EPC5[d6](自ノードインスタンスリストS) PDC5[4] EDT5[01013001] EPC6[d7](自ノードクラスリストS) PDC6[3] EDT6[010130] EPC7[9f](Getプロパティマップ) PDC7[14] EDT7[0d808283898a9d9e9fbfd3d4d6d7]
[Controller]2019/09/25 01:46:36 [192.168.1.10] 108100000ef00105ff0152088001308204010c0100d303000001d4020002d500d60401013001d7030101309f0e0d808283898a9d9e9fbfd3d4d6d7

.[Controller]2019/09/25 01:46:36 <<<<<<<< received
[Frame]2019/09/25 01:46:36 frameReceived: {14 240}
[Frame]2019/09/25 01:46:36 108100000ef00105ff017301d50401013001 EHD[1081] TID[0000] SEOJ[0ef001](ノードプロファイル) DEOJ[05ff01](コントローラ) ESV[INF] OPC[01] EPC0[d5](インスタンスリスト通知) PDC0[4] EDT0[01013001]
[Controller]2019/09/25 01:46:36 [192.168.1.15] 108100000ef00105ff017301d50401013001

.[Controller]2019/09/25 01:46:36 <<<<<<<< received
[Frame]2019/09/25 01:46:36 frameReceived: {14 240}
[Frame]2019/09/25 01:46:36 108100000ef00105ff0152088001308204010c0100d303000001d4020002d500d60401013001d7030101309f0e0d808283898a9d9e9fbfd3d4d6d7 EHD[1081] TID[0000] SEOJ[0ef001](ノードプロファイル) DEOJ[05ff01](コントローラ) ESV[Get_SNA] OPC[08] EPC0[80](動作状態) PDC0[1] EDT0[30] EPC1[82](規格Version情報) PDC1[4] EDT1[010c0100] EPC2[d3](自ノードインスタンス数) PDC2[3] EDT2[000001] EPC3[d4](自ノードクラス数) PDC3[2] EDT3[0002] EPC4[d5](インスタンスリスト通知) PDC4[0] EDT4[] EPC5[d6](自ノードインスタンスリストS) PDC5[4] EDT5[01013001] EPC6[d7](自ノードクラスリストS) PDC6[3] EDT6[010130] EPC7[9f](Getプロパティマップ) PDC7[14] EDT7[0d808283898a9d9e9fbfd3d4d6d7]
[Controller]2019/09/25 01:46:36 [192.168.1.15] 108100000ef00105ff0152088001308204010c0100d303000001d4020002d500d60401013001d7030101309f0e0d808283898a9d9e9fbfd3d4d6d7
..[Frame]2019/09/25 01:46:39 frameReceived: {5 255}
[Frame]2019/09/25 01:46:39 1081000005ff01013001620481008300bb00be00 EHD[1081] TID[0000] SEOJ[05ff01](コントローラ) DEOJ[013001]() ESV[Get] OPC[04] EPC0[81]() PDC0[0] EDT0[] EPC1[83]() PDC1[0] EDT1[] EPC2[bb]() PDC2[0] EDT2[] EPC3[be]() PDC3[0] EDT3[]
[Controller]2019/09/25 01:46:39 >>>>>>>> sendFrame
[Frame]2019/09/25 01:46:39 1081000005ff01013001620481008300bb00be00 EHD[1081] TID[0000] SEOJ[05ff01](コントローラ) DEOJ[013001]() ESV[Get] OPC[04] EPC0[81]() PDC0[0] EDT0[] EPC1[83]() PDC1[0] EDT1[] EPC2[bb]() PDC2[0] EDT2[] EPC3[be]() PDC3[0] EDT3[]
2019/09/25 01:46:39 written: 20
[Controller]2019/09/25 01:46:39 start sendLoop

.
.[Controller]2019/09/25 01:46:39 <<<<<<<< received
[Frame]2019/09/25 01:46:39 frameReceived: {1 48}
[Frame]2019/09/25 01:46:39 1081000001300105ff0172048101418311fe00000860f189306df500000000000000bb011bbe0116 EHD[1081] TID[0000] SEOJ[013001]() DEOJ[05ff01](コントローラ) ESV[Get_Res] OPC[04] EPC0[81]() PDC0[1] EDT0[41] EPC1[83]() PDC1[17] EDT1[fe00000860f189306df500000000000000] EPC2[bb](室内温度計測値) PDC2[1] EDT2[1b] EPC3[be](外気温度計測値) PDC3[1] EDT3[16]
[Frame]2019/09/25 01:46:39 エアコン
[Frame]2019/09/25 01:46:39 Property Code: 81, echonetlite.Data{0x41}
[Frame]2019/09/25 01:46:39 01000001
[Frame]2019/09/25 01:46:39 locationCode: 1000 locationNo: 1
[Frame]2019/09/25 01:46:39 Property Code: 83, echonetlite.Data{0xfe, 0x0, 0x0, 0x8, 0x60, 0xf1, 0x89, 0x30, 0x6d, 0xf5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
[Frame]2019/09/25 01:46:39 メーカコード: echonetlite.Data{0x0, 0x0, 0x8} メーカID: echonetlite.Data{0x60, 0xf1, 0x89, 0x30, 0x6d, 0xf5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
[Frame]2019/09/25 01:46:39 Property Code: bb, echonetlite.Data{0x1b}
[Frame]2019/09/25 01:46:39 室温:27℃
[Frame]2019/09/25 01:46:39 Property Code: be, echonetlite.Data{0x16}
[Frame]2019/09/25 01:46:39 外気温:22℃
[Controller]2019/09/25 01:46:39 [192.168.1.15] 1081000001300105ff0172048101418311fe00000860f189306df500000000000000bb011bbe0116
[Controller]2019/09/25 01:46:39 <<<<<<<< received
[Frame]2019/09/25 01:46:39 frameReceived: {1 48}
[Frame]2019/09/25 01:46:39 1081000001300105ff0172048101088311fe00000800aefac9243500000000000000bb0118be0117 EHD[1081] TID[0000] SEOJ[013001]() DEOJ[05ff01](コントローラ) ESV[Get_Res] OPC[04] EPC0[81]() PDC0[1] EDT0[08] EPC1[83]() PDC1[17] EDT1[fe00000800aefac9243500000000000000] EPC2[bb](室内温度計測値) PDC2[1] EDT2[18] EPC3[be](外気温度計測値) PDC3[1] EDT3[17]
[Frame]2019/09/25 01:46:39 エアコン
[Frame]2019/09/25 01:46:39 Property Code: 81, echonetlite.Data{0x8}
[Frame]2019/09/25 01:46:39 00001000
[Frame]2019/09/25 01:46:39 locationCode: 1 locationNo: 0
[Frame]2019/09/25 01:46:39 Property Code: 83, echonetlite.Data{0xfe, 0x0, 0x0, 0x8, 0x0, 0xae, 0xfa, 0xc9, 0x24, 0x35, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
[Frame]2019/09/25 01:46:39 メーカコード: echonetlite.Data{0x0, 0x0, 0x8} メーカID: echonetlite.Data{0x0, 0xae, 0xfa, 0xc9, 0x24, 0x35, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
[Frame]2019/09/25 01:46:39 Property Code: bb, echonetlite.Data{0x18}
[Frame]2019/09/25 01:46:39 室温:24℃
[Frame]2019/09/25 01:46:39 Property Code: be, echonetlite.Data{0x17}
[Frame]2019/09/25 01:46:39 外気温:23℃
[Controller]2019/09/25 01:46:39 [192.168.1.10] 1081000001300105ff0172048101088311fe00000800aefac9243500000000000000bb0118be0117
```
