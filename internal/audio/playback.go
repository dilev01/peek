package audio

import (
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	bwav "github.com/faiface/beep/wav"
)

type Player interface {
	Play(filepath string) error
	Stop() error
}

type BeepPlayer struct {
	initialized bool
}

func NewBeepPlayer() *BeepPlayer {
	return &BeepPlayer{}
}

func (p *BeepPlayer) Play(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	streamer, format, err := bwav.Decode(f)
	if err != nil {
		f.Close()
		return err
	}
	if !p.initialized {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		p.initialized = true
	}
	done := make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		streamer.Close()
		close(done)
	})))
	<-done
	return nil
}

func (p *BeepPlayer) Stop() error {
	speaker.Clear()
	return nil
}
