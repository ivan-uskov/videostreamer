package transport

import (
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/vpx"
	"github.com/pion/mediadevices/pkg/prop"
	log "github.com/sirupsen/logrus"
	"strings"

	_ "github.com/pion/mediadevices/pkg/driver/camera"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
)

type MediaDevicesConfig struct {
	CameraID     string `envconfig:"camera_id"`
	CameraFormat string `envconfig:"camera_format"`
	BitRate      int    `envconfig:"bit_rate"`
	VideoWidth   int    `envconfig:"video_width"`
	VideoHeight  int    `envconfig:"video_height"`
}

func GetCodecSelector(conf MediaDevicesConfig) *mediadevices.CodecSelector {
	vpxParams, err := vpx.NewVP8Params()
	if err != nil {
		log.WithError(err).Fatal("vp8 error")
	}
	vpxParams.BitRate = conf.BitRate

	return mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&vpxParams),
	)
}

func GetMediaDeviceStream(codec *mediadevices.CodecSelector, conf MediaDevicesConfig) mediadevices.MediaStream {
	device, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraints *mediadevices.MediaTrackConstraints) {
			constraints.DeviceID = prop.String(getVideoDeviceId(conf.CameraID))
			constraints.FrameFormat = prop.FrameFormat(conf.CameraFormat)
			constraints.Width = prop.Int(conf.VideoWidth)
			constraints.Height = prop.Int(conf.VideoHeight)
		},
		Codec: codec,
	})
	if err != nil {
		log.WithError(err).Fatal("get user media error")
	}

	return device
}

func getVideoDeviceId(label string) string {
	devices := mediadevices.EnumerateDevices()
	log.WithField("dd", devices).Debug("media devices")

	for _, d := range devices {
		if strings.Contains(d.Label, label) {
			log.WithField("camera", d).Info("user camera")
			return d.DeviceID
		}
	}

	log.WithField("device", label).Fatalf("device not found")
	return ""
}

type Listener func(t mediadevices.Track)

func BindStreamTrackListeners(s mediadevices.MediaStream, listener Listener) {
	for _, track := range s.GetTracks() {
		track.OnEnded(func(err error) {
			log.WithError(err).WithField("track", track.ID()).Error("Track ended with error")
		})
		listener(track)
	}
}
