package frame

const MaxUserId = 10000000000

func GenAppUid(appid int32, uid uint64) uint64 {
	if uid > MaxUserId {
		return uid
	}
	if !defFrameOption.EnableAppUid {
		return uid
	}
	return uint64(appid)*MaxUserId + uid
}
func GetUidFromAppUid(uid uint64) uint64 {
	return uid % MaxUserId
}

func GetAppidFromAppUid(uid uint64) int32 {
	return int32(uid / MaxUserId)
}
