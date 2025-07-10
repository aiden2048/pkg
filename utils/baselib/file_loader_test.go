package baselib

import (
	"log"
	"testing"
)

func TestFileNameWithFullPath(t *testing.T) {
	log.Println(FileLocation("map.go"))
	log.Println(FileLocation("etcd.toml"))
	log.Println(FileLocation("x.toml"))
}

func TestDecodeToml(t *testing.T) {
	type LogConfig struct {
		logLevel         int
		Level            string
		TraceAllPlayer   bool
		TraceGlobePlayer bool
	}

	type RunConfig struct {
		Fe        uint32
		Mode      string
		RouteType string
		mode      int
	}

	// ProfileConfig Dump CPU/Mem/Stack信息时配置
	type ProfileConfig struct {
		PrintStack bool
	}

	type FrameConfig struct {
		Log     LogConfig
		Frame   RunConfig
		Profile ProfileConfig
	}

	c := &FrameConfig{}
	if err := DecodeToml("frame.toml", &c); err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", *c)
}
