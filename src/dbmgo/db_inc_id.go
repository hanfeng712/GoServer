package dbmgo

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
)

var (
	g_inc_id_map   = make(map[string]uint32)
	g_inc_id_mutex sync.Mutex
)

type nameId struct {
	Name string `bson:"_id"`
	ID   uint32
}

func _init_inc_ids() {
	var lst []nameId
	FindAll("IncId", nil, &lst)
	g_inc_id_mutex.Lock()
	for _, v := range lst {
		g_inc_id_map[v.Name] = v.ID
	}
	g_inc_id_mutex.Unlock()
}
func GetNextIncId(key string) uint32 {
	g_inc_id_mutex.Lock()
	ret := g_inc_id_map[key] + 1 //Debug：实际包含三步：读出、+1、写入，必原子的完成，才可保证每次返回不同id；sync.Map仅保障了读写安全性
	g_inc_id_map[key] = ret
	g_inc_id_mutex.Unlock()
	if ret == 1 {
		Insert("IncId", nameId{key, 1})
	} else {
		UpdateId("IncId", key, bson.M{"$set": bson.M{"id": ret}})
	}
	return ret
}
