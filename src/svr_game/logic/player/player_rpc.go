/***********************************************************************
* @ 与玩家强绑定的rpc，比对net_rpc.go
* @ brief
	1、将原生rpc的参数转换为 player *TPlayer

	2、拦截原生网络rpc，处理通用部分得到*TPlayer，再转入PlayerRpc

* @ author zhoumf
* @ date 2018-3-23
***********************************************************************/
package player

import (
	"common"
	"common/compress"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	mhttp "http"
	"net/http"
	"netConfig"
	"sync/atomic"
	"tcp"
)

type PlayerRpc func(req, ack *common.NetPack, this *TPlayer)

var G_PlayerHandleFunc [enum.RpcEnumCnt]PlayerRpc

// 访问玩家数据的消息，要求该玩家已在缓存中，否则不处理
//【Notice：登录、创建角色，可做成普通rpc，用以建立玩家缓存】
func RegPlayerRpc(list map[uint16]PlayerRpc) {
	for k, v := range list {
		G_PlayerHandleFunc[k] = v
	}
	tcp.RegHandlePlayerRpc(_HandlePlayerRpc1)   //tcp 直连
	mhttp.RegHandlePlayerRpc(_HandlePlayerRpc2) //http 直连
}
func DoPlayerRpc(this *TPlayer, rpcId uint16, req, ack *common.NetPack) bool {
	if msgFunc := G_PlayerHandleFunc[rpcId]; msgFunc != nil {
		atomic.SwapUint32(&this._idleMin, 0)
		msgFunc(req, ack, this)
		return true
	}
	gamelog.Error("PlayerMsg(%d) Not Regist", rpcId)
	return false
}

// ------------------------------------------------------------
// - tcp 直连 player rpc
// - 将原生tcpRpc的 "conn *tcp.TCPConn" 参数转换为 "player *TPlayer"
func _HandlePlayerRpc1(req, ack *common.NetPack, conn *tcp.TCPConn) bool {
	if msgFunc := G_PlayerHandleFunc[req.GetOpCode()]; msgFunc != nil {
		if player, ok := conn.UserPtr.(*TPlayer); ok {
			atomic.SwapUint32(&player._idleMin, 0)
			msgFunc(req, ack, player)
		}
		return true
	}
	return false
}

// ------------------------------------------------------------
// - http 直连 player rpc
func _HandlePlayerRpc2(w http.ResponseWriter, r *http.Request) {
	//! 接收信息
	req := mhttp.ReadRequest(r)
	if req == nil {
		return
	}
	defer req.Free()

	//! 创建回复
	ack := common.NewNetPackCap(128)
	msgId := req.GetOpCode()
	accountId := req.GetReqIdx()
	defer ack.Free()
	//defer func() {//库已经有recover了，见net/http/server.go:1918
	//	if r := recover(); r != nil {
	//		gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
	//	}
	//	ack.Free()
	//}()
	//FIXME: 验证消息安全性，防改包
	//FIXME: http通信中途安全性不够，能修改client net pack里的uid，进而操作别人数据
	//FIXME: 账号服登录验证后下发给client的token，client应该保留，附在每个HttpReq里，防止恶意窜改他人数据

	if conf.IsDebug && msgId != enum.Rpc_game_heart_beat {
		gamelog.Debug("HttpMsg:%d, len:%d, uid:%d", msgId, req.Size(), accountId)
	}
	if player := BeforeRecvHttpMsg(accountId); player != nil {
		if DoPlayerRpc(player, msgId, req, ack) {
			AfterRecvHttpMsg(player, ack)
			compress.CompressTo(ack.Data(), w)
		}
	} else {
		gamelog.Debug("Player(%d) isn't online", accountId)
	}
}

// ------------------------------------------------------------
// - 网关转发的玩家消息
func Rpc_recv_player_msg(req, ack *common.NetPack, conn *tcp.TCPConn) {
	rpcId := req.ReadUInt16()
	accountId := req.ReadUInt32()

	gamelog.Debug("PlayerMsg:%d", rpcId)

	if player := FindAccountId(accountId); player != nil {
		DoPlayerRpc(player, rpcId, req, ack)
	} else {
		gamelog.Debug("Player(%d) isn't online", accountId)
	}
}

// ------------------------------------------------------------
// - 与其它玩家交互(可能位于其它节点，能通知到别人客户端)
func CallRpcPlayer(accountId uint32, msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	if msgFunc := G_PlayerHandleFunc[msgId]; msgFunc != nil {
		if player := FindAccountId(accountId); player != nil {
			req := common.NewNetPackCap(64)
			ack := common.NewNetPackCap(64)
			req.SetOpCode(msgId)
			sendFun(req)
			msgFunc(req, ack, player)
			if recvFun != nil {
				recvFun(ack)
			}
			req.Free()
			ack.Free()
			return
		}
	}
	netConfig.CallRpcGateway(accountId, msgId, sendFun, recvFun)
}
