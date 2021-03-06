package account

import (
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	g_name_cache sync.Map //map[string]*TAccount
	g_aid_cache  sync.Map //map[uint32]*TAccount
)

func InitDB() {
	//只载入一个月内登录过的
	var list []TAccount
	dbmgo.FindAll(kDBTable, bson.M{"logintime": bson.M{"$gt": time.Now().Unix() - 30*24*3600}}, &list)
	for i := 0; i < len(list); i++ {
		list[i].init()
		AddCache(&list[i])
	}
	println("load active account form db: ", len(list))
}
func AddNewAccount(name, passwd string) *TAccount {
	account := _NewAccount()
	if dbmgo.Find(kDBTable, "name", name, account) {
		return nil
	}
	account.Name = name
	account.SetPasswd(passwd)
	account.CreateTime = time.Now().Unix()
	account.AccountID = dbmgo.GetNextIncId("AccountId")

	if dbmgo.InsertSync(kDBTable, account) {
		AddCache(account)
		return account
	}
	return nil
}
func GetAccountByName(name string) *TAccount {
	if v, ok := g_name_cache.Load(name); ok {
		return v.(*TAccount)
	} else {
		account := _NewAccount()
		if dbmgo.Find(kDBTable, "name", name, account) {
			AddCache(account)
			return account
		}
	}
	return nil
}
func GetAccountById(accountId uint32) *TAccount {
	if v, ok := g_aid_cache.Load(accountId); ok {
		return v.(*TAccount)
	} else {
		account := _NewAccount()
		if dbmgo.Find(kDBTable, "_id", accountId, account) {
			AddCache(account)
			return account
		}
	}
	return nil
}

// -------------------------------------
//! 辅助函数
func AddCache(account *TAccount) {
	g_name_cache.Store(account.Name, account)
	g_aid_cache.Store(account.AccountID, account)
}
func DelCache(account *TAccount) {
	g_name_cache.Delete(account.Name)
	g_aid_cache.Delete(account.AccountID)
}
