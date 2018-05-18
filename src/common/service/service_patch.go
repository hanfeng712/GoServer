package service

// -------------------------------------
// -- 花一段时长，遍历完所有对象
type ServicePatch struct {
	callback  func(interface{})
	timeWait  int // msec
	kTimeAll  int // msec
	runPos    int
	obj_lst   []interface{}
	writeChan chan ServiceObj
}

func NewServicePatch(fun func(interface{}), timeAllMsec int) *ServicePatch {
	ptr := new(ServicePatch)
	ptr.callback = fun
	ptr.kTimeAll = timeAllMsec
	ptr.writeChan = make(chan ServiceObj, 64)
	return ptr
}
func (self *ServicePatch) UnRegister(pObj interface{}) { self.writeChan <- ServiceObj{pObj, false} }
func (self *ServicePatch) Register(pObj interface{})   { self.writeChan <- ServiceObj{pObj, true} }
func (self *ServicePatch) RunSevice(timelapse int, timenow int64) {
	for {
		select {
		case data := <-self.writeChan:
			if data.isReg {
				self._doRegister(data.pObj)
			} else {
				self._doUnRegister(data.pObj)
			}
		default:
			self._runSevice(timelapse)
			return
		}
	}
}
func (self *ServicePatch) _doRegister(pObj interface{}) { self.obj_lst = append(self.obj_lst, pObj) }
func (self *ServicePatch) _doUnRegister(pObj interface{}) {
	for i, ptr := range self.obj_lst {
		if ptr == pObj {
			self.obj_lst = append(self.obj_lst[:i], self.obj_lst[i+1:]...)
			if i < self.runPos {
				self.runPos--
			} else if self.runPos >= len(self.obj_lst) {
				self.runPos = 0
			}
			return
		}
	}
}
func (self *ServicePatch) _runSevice(timelapse int) {
	if len(self.obj_lst) <= 0 {
		return
	}
	//! 单位时长里要处理的个数，可能大于列表中obj总数，比如服务器卡顿很久，得追帧
	self.timeWait += timelapse
	runCnt := self.timeWait * len(self.obj_lst) / self.kTimeAll
	if runCnt == 0 {
		return
	}
	//! 处理一个的时长
	temp := self.kTimeAll / len(self.obj_lst)
	//! 更新等待时间(须小于"处理一个的时长")：对"处理一个的时长"取模(除法的非零保护)
	if temp > 0 {
		self.timeWait %= temp
	} else {
		self.timeWait = 0
	}

	for i := 0; i < runCnt; i++ {
		obj := self.obj_lst[self.runPos]
		if self.runPos++; self.runPos >= len(self.obj_lst) { //到末尾了，回到队头
			self.runPos = 0
		}
		self.callback(obj)
	}
}
