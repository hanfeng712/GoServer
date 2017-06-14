package main

import (
	"common"
	"conf"
	"dbmgo"
	"gamelog"
	"netConfig"
	"strconv"

	"svr_game/center"
	"svr_game/cross"
	"svr_game/logic"
	"svr_game/logic/msg"
	"svr_game/logic/player"
	"svr_game/sdk"
)

func main() {
	//初始化日志系统
	gamelog.InitLogger("game")
	gamelog.SetLevel(0)

	//设置mongodb的服务器地址
	dbmgo.Init(conf.GameDbAddr, conf.GameDbName)

	//开启控制台窗口，可以接受一些调试命令
	common.StartConsole()
	common.RegConsoleCmd("setloglevel", HandCmd_SetLogLevel)
	common.RegConsoleCmd("MakeFriends", HandCmd_MakeFriends)

	InitConf()

	go logic.MainLoop()

	gamelog.Warn("----Game Server Start-----")
	if netConfig.CreateNetSvr("game", 1) == false {
		gamelog.Error("----Game NetSvr Failed-----")
	}
}
func HandCmd_SetLogLevel(args []string) bool {
	level, err := strconv.Atoi(args[1])
	if err != nil {
		gamelog.Error("HandCmd_SetLogLevel => Invalid param:%s", args[1])
		return false
	}
	gamelog.SetLevel(level)
	return true
}
func HandCmd_MakeFriends(args []string) bool {
	pid1, err1 := strconv.Atoi(args[1])
	pid2, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		gamelog.Error("HandCmd_MakeFriends => Invalid param:%s, %s", args[1], args[2])
		return false
	}
	player1 := player.FindWithDB_PlayerId(uint32(pid1))
	player2 := player.FindWithDB_PlayerId(uint32(pid2))
	if player1 != nil && player2 != nil {
		player1.AsyncNotify(func(player *player.TPlayer) {
			player.Friend.AddFriend(player2.PlayerID, player2.Name)
		})
		player2.AsyncNotify(func(player *player.TPlayer) {
			player.Friend.AddFriend(player1.PlayerID, player1.Name)
		})
		return true
	}
	return false
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &netConfig.G_SvrNetCfg,
		"rpc":      &common.G_RpcCsv,
	}
	common.LoadAllCsv()

	netConfig.RegTcpHandler(map[string]netConfig.TcpHandle{
		//! Cross
		"rpc_echo":       cross.Rpc_Echo,
		"rpc_battle_ack": player.Rpc_Battle_Ack,
	})
	netConfig.RegHttpSystemHandler(map[string]netConfig.HttpHandle{
		//! Center
		"login_token": center.Handle_Login_Token,
		//! SDK
		"create_recharge_order": sdk.Handle_Create_Recharge_Order,
		"sdk_recharge_success":  sdk.Handle_Recharge_Success,
	})
	netConfig.RegHttpPlayerHandler(map[string]netConfig.HttpPlayerHandle{
		//! Client
		"battle_echo":       msg.Rpc_Client2Battle_Echo,
		"rpc_test_mongodb":  msg.Rpc_test_mongodb,
		"rpc_login":         player.Rpc_Player_Login,
		"rpc_logout":        player.Rpc_Player_Logout,
		"rpc_player_create": player.Rpc_Player_Create,
		"rpc_battle_begin":  player.Rpc_Battle_Begin,
		"rpc_friend_list":   player.Rpc_Friend_List,
		"rpc_friend_apply":  player.Rpc_Friend_Apply,
		"rpc_friend_agree":  player.Rpc_Friend_Agree,
		"rpc_friend_refuse": player.Rpc_Friend_Refuse,
		"rpc_friend_del":    player.Rpc_Friend_Del,
	})
	netConfig.G_Before_Recv_Player_Http = player.BeforeRecvHttpMsg
	netConfig.G_After_Recv_Player_Http = player.AfterRecvHttpMsg
}
