package BW800Tcp

import (
	"errors"
	"fmt"
	"github.com/wumingyu12/golang_YuGoTool"
)

type RecordStructList struct {
	List []RecordStruct
}
type RecordStruct struct {
	Addr   string //uint32 //我们自己加上去的代表这条记录是哪个控制器地址产生的,因为odoo里面用的变量是char所以我们也用string,odoo不用interger是避免溢出
	PigNum string //uint32 //猪耳标
	Date   string // C=04 记录    : 日期，如 0x27182 = 160130 = 16年1月30日
	// C=05记录明细: 时标 从1970年1月1日0点0分0秒 到现在的秒数
	// 时标为0或者0xffffffff时，记录无效
	Amount uint16 //总量
	Actual uint16 //实际
	Sum    uint32 //校验和
}

/************************************
用16长的字节实例化这个结构体
使用这个方法前要保证已经实例化了
dayOrdetail  可以输入day或者detail
用来指示实例化的是日下料记录还是详细下料记录，这两个记录的日期表示方法是不同的
 **********************************/
func (this *RecordStruct) Reflash(b []byte, dayOrdetail string) error {
	//fmt.Println(len(b))
	//b := Fun_handle_message(mes) //去掉包头,包尾校验
	if len(b) != 16 {
		return errors.New("存入Bw800Para方法Reflash的参数去掉包头后不是16个字节")
	}
	num := 0                                                                                           //指示b的下标\
	this.PigNum = fmt.Sprintf("%d", YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num])) //odoo那边是char所以我们也用string
	if len(this.PigNum) < STRING_LONG {                                                                //如果长度不足
		var temp string
		for i := 0; i < STRING_LONG-len(this.PigNum); i++ {
			temp = temp + "0"
		}
		this.PigNum = temp + this.PigNum
	}
	num = num + 4
	//如果日期解释是作为日下料记录解析
	if dayOrdetail == "day" {
		da := YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num])
		//160315 转为16-03-15
		//小于10 补零
		shiwei := fmt.Sprintf("%d", (da%10000)/100)
		if (da%10000)/100 < 10 {
			shiwei = "0" + shiwei
		}
		gewei := fmt.Sprintf("%d", da%100)
		if da%100 < 10 {
			gewei = "0" + gewei
		}
		this.Date = fmt.Sprintf("%d-%s-%s", da/10000, shiwei, gewei)
		//fmt.Printf("%d-%s-%s\n", da/10000, shiwei, gewei)
	}
	num = num + 4
	this.Amount = YuGoTool.Twobyte_to_uint16(b[num+1], b[num])
	num = num + 2
	this.Actual = YuGoTool.Twobyte_to_uint16(b[num+1], b[num])
	num = num + 2
	this.Sum = YuGoTool.Fourbyte_to_uint32(b[num+3], b[num+2], b[num+1], b[num])
	return nil
}
