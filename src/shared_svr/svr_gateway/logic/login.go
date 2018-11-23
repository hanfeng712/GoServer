package logic

import (
	"common"
	"generate_out/err"
	"generate_out/rpc/enum"
	"http"
	"netConfig"
	"netConfig/meta"
	"sync"
	"tcp"
)

func Rpc_gateway_login(req, ack *common.NetPack, client *tcp.TCPConn) {
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()

	if CheckLoginToken(accountId, token) {
		client.UserPtr = accountId
		AddClientConn(accountId, client)
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Token_verify_err)
	}
}

// RelayPlayerMsg处理的玩家相关rpc（rpc参数是this *TPlayer）
// 登录之前，游戏服尚无玩家数据，所以登录、创建是单独抽离的
func Rpc_gateway_relay_game_login(req, ack *common.NetPack, client *tcp.TCPConn) {
	if accountId, ok := client.UserPtr.(uint32); ok {
		if pConn := GetGameConn(accountId); pConn != nil {
			oldReqKey := req.GetReqKey()
			pConn.CallRpc(enum.Rpc_game_login, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteBuf(req.LeftBuf())
			}, func(backBuf *common.NetPack) {
				//异步回调，不能直接用ack
				backBuf.SetReqKey(oldReqKey)
				client.WriteMsg(backBuf)
			})
		}
	}
}
func Rpc_gateway_relay_game_create_player(req, ack *common.NetPack, client *tcp.TCPConn) {
	if accountId, ok := client.UserPtr.(uint32); ok {
		if pConn := GetGameConn(accountId); pConn != nil {
			oldReqKey := req.GetReqKey()
			pConn.CallRpc(enum.Rpc_game_create_player, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteBuf(req.LeftBuf())
			}, func(backBuf *common.NetPack) {
				//异步回调，不能直接用ack
				backBuf.SetReqKey(oldReqKey)
				client.WriteMsg(backBuf)
			})
		}
	}
}

// ------------------------------------------------------------
// -- 后台账号验证
var g_login_token sync.Map //<accountId, token>

func Rpc_gateway_login_token(req, ack *common.NetPack, conn *tcp.TCPConn) {
	token := req.ReadUInt32()
	accountId := req.ReadUInt32()
	gameSvrId := req.ReadInt()

	g_login_token.Store(accountId, token)

	AddGameConn(accountId, gameSvrId) //设置此玩家的game路由

	//取游戏服在线人数，发给登录服
	netConfig.CallRpcGame(gameSvrId, enum.Rpc_game_get_player_cnt, func(buf *common.NetPack) {
	}, func(backBuf *common.NetPack) {
		cnt := backBuf.ReadInt32()
		_NotifyPlayerCnt(gameSvrId, cnt)
	})
}
func CheckLoginToken(accountId, token uint32) bool {
	if value, ok := g_login_token.Load(accountId); ok {
		return token == value
	}
	return false
}

// ------------------------------------------------------------
// -- 游戏服在线人数
func _NotifyPlayerCnt(gameSvrId int, cnt int32) {
	ids, _ := meta.GetModuleIDs("login", netConfig.G_Local_Meta.Version)
	for _, id := range ids {
		if addr := netConfig.GetHttpAddr("login", id); addr != "" {
			http.CallRpc(addr, enum.Rpc_login_set_player_cnt, func(buf *common.NetPack) {
				buf.WriteInt(gameSvrId)
				buf.WriteInt32(cnt)
			}, nil)
		}
	}
}
