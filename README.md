```
    if err := frame.InitConfig("userRiskData", &frame.FrameOption{
		Port: commonConst.PortUserRiskData,
	}); err != nil {
		log.Fatalf("InitConfig error:%s", err.Error())
		return
	}
    defer frame.Stop()
	frame.Run(handler.RegisterHandlers)
```