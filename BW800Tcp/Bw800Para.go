package BW800Tcp

import (
	"errors"
	"github.com/wumingyu12/golang_YuGoTool"
)

//参数结构体
type Bw800Para struct {
	TotalTable  []uint16 //总量曲线 单位10克 1234 表示 12.34kg  20个数据
	TimeTStart  []uint16 //饲喂时间开始 10个数据
	TimeTEnd    []uint16 //饲喂时间结束 10个数据
	TimeTAmount []uint16 //饲喂总量 10个数据
	CalTable    []uint16 //校正表 -- g 10个数据
	CanOver     uint16   //允许超出量                -- 10%
	XlPer       uint16   //每次下料多少克 -- 100g
	WaterAuto   uint16   //是否自动下水
	WaterTime   uint16   //下水时间s
	WaterSpace  uint16   //下水间隔s
	UsePass     uint16   //是否使用密码 0 - 不使用 1 - 使用
	Addr        uint32   //控制器地址
	RouteName   uint32   //路由器名称
	RoutePass   uint32   //路由器密码
	ServerIP    []uint32 //服务器IP地址 4个
	ServerPort  uint32   //服务器端口
	Stage       uint16   //0 - 断奶 1 - 产前 2 - 哺乳
	Day         uint16   //第几日
	HasEat      uint16   //今天已吃 -- g
	EatDelay    uint16   //要求喂食延迟
	Sons        uint16   //仔猪头数
	Mday        uint8    //保存日
	Rev         uint8    //保留
	Password    uint32   //用户密码
	Sum         uint32   //校验和
}

/************************************
	用184长的字节实例化这个结构体
	使用这个方法前要保证已经实例化了
 **********************************/
func (bw *Bw800Para) Reflash(mes []byte) error {
	b := Fun_handle_message(mes) //去掉包头,包尾校验
	if len(b) != 184 {
		return errors.New("存入Bw800Para方法Reflash的参数去掉包头后不是184个字节")
	}
	num := 0 //指示b的下标
	//总量曲线 单位10克 1234 表示 12.34kg  20个数据
	bw.TotalTable = make([]uint16, 20)
	for i := 0; i < 20; i++ {
		bw.TotalTable[i] = YuGoTool.Twobyte_to_uint16(b[num+1], b[num])
		num = num + 2
	}
	//饲喂时间开始 10个数据
	bw.TimeTStart = make([]uint16, 10)
	for i := 0; i < 10; i++ {
		bw.TimeTStart[i] = YuGoTool.Twobyte_to_uint16(b[num+1], b[num])
		num = num + 2
	}
	//饲喂时间结束 10个数据
	bw.TimeTEnd = make([]uint16, 10)
	for i := 0; i < 10; i++ {
		bw.TimeTEnd[i] = YuGoTool.Twobyte_to_uint16(b[num+1], b[num])
		num = num + 2
	}
	//饲喂总量 10个数据
	bw.TimeTAmount = make([]uint16, 10)
	for i := 0; i < 10; i++ {
		bw.TimeTAmount[i] = YuGoTool.Twobyte_to_uint16(b[num+1], b[num])
		num = num + 2
	}
	//校正表 -- g 10个数据
	bw.CalTable = make([]uint16, 10)
	for i := 0; i < 10; i++ {
		bw.CalTable[i] = YuGoTool.Twobyte_to_uint16(b[num+1], b[num])
		num = num + 2
	}
	bw.CanOver = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //允许超出量                -- 10%
	num = num + 2
	bw.XlPer = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //每次下料多少克 -- 100g
	num = num + 2
	bw.WaterAuto = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //是否自动下水
	num = num + 2
	bw.WaterTime = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //下水时间s
	num = num + 2
	bw.WaterSpace = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //下水间隔s
	num = num + 2
	bw.UsePass = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //是否使用密码 0 - 不使用 1 - 使用
	num = num + 2
	bw.Addr = YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num]) //控制器地址
	num = num + 4
	bw.RouteName = YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num]) //路由器名称
	num = num + 4
	bw.RoutePass = YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num]) //路由器密码
	num = num + 4
	//服务器IP地址 4个
	bw.ServerIP = make([]uint32, 4)
	for i := 0; i < 4; i++ {
		bw.ServerIP[i] = YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num])
		num = num + 4
	}
	bw.ServerPort = YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num]) //服务器端口
	num = num + 4
	bw.Stage = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //0 - 断奶 1 - 产前 2 - 哺乳
	num = num + 2
	bw.Day = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //第几日
	num = num + 2
	bw.HasEat = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //今天已吃 -- g
	num = num + 2
	bw.EatDelay = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //要求喂食延迟
	num = num + 2
	bw.Sons = YuGoTool.Twobyte_to_uint16(b[num+1], b[num]) //仔猪头数
	num = num + 2
	bw.Mday = uint8(b[num]) //保存日
	num = num + 1
	bw.Rev = uint8(b[num]) //保留
	num = num + 1
	bw.Password = YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num]) //用户密码
	num = num + 4
	bw.Sum = YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num]) //校验和

	return nil
}
