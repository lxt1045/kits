package channel

//"sync"

//IChanN 用于IO和逻辑层之间传递数据可以用原生封装，也可以新建
//要求：发送不阻塞，而且缓冲满了之后，要求覆盖旧的数据； 接收可选择是否阻塞
type IChanN interface {
	Send(data interface{}) (full bool, closed bool)
	SendN(datas []interface{}) (n int, closed bool)
	Recv(n int, block int) (result []interface{}, closed bool)
	Read(datas []interface{}, block int) (n int, closed bool)
	Close()
}
