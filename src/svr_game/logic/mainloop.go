package logic

import (
	"common"
	"common/timer"
	"conf"
	"svr_game/logic/player"
	"tcp"
	"time"
)

func MainLoop() {
	timerChan := timer.NewTimerChan(40960)
	InitTimeUpdate(timerChan)

	timeNow, timeOld, timeElapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0
	for {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		timeElapse = int(timeNow - timeOld)

		player.G_ServiceMgr.RunAllService(timeElapse, timeNow)

		tcp.G_RpcQueue.Update()

		timerChan.Update()

		if timeElapse < conf.FPS_GameSvr {
			time.Sleep(time.Duration(conf.FPS_GameSvr-timeElapse) * time.Millisecond)
		}
	}
}
func Rpc_net_error(req, ack *common.NetPack, conn *tcp.TCPConn) {
}
