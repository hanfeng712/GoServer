/***********************************************************************
* @ 玩家数据
* @ brief
	1、数据散列模块化，按业务区分成块，各自独立处理，如：TBaseMoudle、TMailMoudle
	2、可调用DB【同步读单个模块】，编辑后再【异步写】

* @ 访问离线玩家
	1、设想把TPlayer里的数据块部分全定义为指针，各模块分别做个缓存表(online list、offline list)
	2、但觉得有些设计冗余，缓存这种事情，应该交给DBCache系统做，业务层不该负责这事儿

* @ author zhoumf
* @ date 2017-4-22
***********************************************************************/
package player

import (
	"dbmgo"
	"sync"

	"svr_game/logic/mail"
)

type PlayerMoudle interface {
	InitWriteDB(id uint32)
	LoadFromDB(id uint32)
	OnLogin()
	OnLogout()
}
type TBaseMoudle struct {
	PlayerID   uint32 `bson:"_id"`
	AccountID  uint32
	Name       string
	LoginTime  int64
	LogoutTime int64
}
type TPlayer struct {
	//db data
	Base TBaseMoudle
	Mail mail.TMailMoudle
	//temp data
	mutex   sync.Mutex
	moudles []PlayerMoudle
}

func NewPlayer(accountId uint32, id uint32, name string) *TPlayer {
	player := new(TPlayer)
	//! regist
	player.moudles = []PlayerMoudle{
		&player.Mail,
	}
	player.Base.AccountID = accountId
	player.Base.PlayerID = id
	player.Base.Name = name
	if err := dbmgo.InsertSync("Player", &player.Base); err != nil {
		player.InitWriteDB()
		return player
	}
	return nil
}
func (self *TPlayer) InitWriteDB() {
	for _, v := range self.moudles {
		v.InitWriteDB(self.Base.PlayerID)
	}
}
func (self *TPlayer) LoadAllFromDB(id uint32) bool {
	if ok := dbmgo.Find("Player", "_id", id, &self.Base); ok {
		for _, v := range self.moudles {
			v.LoadFromDB(id)
		}
		return true
	}
	return false
}
func (self *TPlayer) OnLogin() {
	for _, v := range self.moudles {
		v.OnLogin()
	}
}
func (self *TPlayer) OnLogout() {
	for _, v := range self.moudles {
		v.OnLogout()
	}
}