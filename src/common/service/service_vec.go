package service

// -------------------------------------
// -- 周期性调用传入的函数
type TimePair struct {
	time int64
	obj  interface{}
}
type ServiceVec struct {
	callback  func(interface{})
	runPos    int
	cdMs      int
	obj_lst   []TimePair
	writeChan chan ServiceObj
}

func NewServiceVec(fun func(interface{}), cdMs int) *ServiceVec {
	ptr := new(ServiceVec)
	ptr.cdMs = cdMs
	ptr.callback = fun
	ptr.writeChan = make(chan ServiceObj, 64)
	return ptr
}
func (self *ServiceVec) UnRegister(pObj interface{}) { self.writeChan <- ServiceObj{pObj, false} }
func (self *ServiceVec) Register(pObj interface{})   { self.writeChan <- ServiceObj{pObj, true} }
func (self *ServiceVec) RunSevice(timelapse int, timenow int64) {
	for {
		select {
		case data := <-self.writeChan:
			if data.isReg {
				self._doRegister(data.pObj)
			} else {
				self._doUnRegister(data.pObj)
			}
		default:
			self._runSevice(timenow)
			return
		}
	}
}
func (self *ServiceVec) _doRegister(pObj interface{}) {
	self.obj_lst = append(self.obj_lst, TimePair{0, pObj})
}
func (self *ServiceVec) _doUnRegister(pObj interface{}) {
	for i, it := range self.obj_lst {
		if it.obj == pObj {
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
func (self *ServiceVec) _runSevice(timenow int64) {
	for self.runPos < len(self.obj_lst) {
		if it := &self.obj_lst[self.runPos]; it.time <= timenow {
			runObj := it.obj
			if self.runPos++; self.runPos >= len(self.obj_lst) { //到末尾了，回到队头
				self.runPos = 0
			}
			if self.callback(runObj); it.obj == runObj { //里头可能删掉自己，it指向改变
				it.time = timenow + int64(self.cdMs)
			}
		} else {
			break
		}
	}
}
