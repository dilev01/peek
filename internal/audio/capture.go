package audio

import (
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	SampleRate      = 16000
	NumChannels     = 1
	BitDepth        = 16
	FramesPerBuffer = 512
)

type Recorder interface {
	Start() error
	Stop() ([]int16, error)
	IsRecording() bool
}

type PortAudioRecorder struct {
	mu        sync.Mutex
	stream    *portaudio.Stream
	buffer    []int16
	recording bool
	samples   []int16
}

func NewPortAudioRecorder() *PortAudioRecorder {
	return &PortAudioRecorder{
		buffer: make([]int16, FramesPerBuffer),
	}
}

func (r *PortAudioRecorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.recording {
		return nil
	}
	stream, err := portaudio.OpenDefaultStream(1, 0, SampleRate, len(r.buffer), r.buffer)
	if err != nil {
		return err
	}
	r.stream = stream
	r.samples = nil
	r.recording = true
	if err := stream.Start(); err != nil {
		return err
	}
	go r.readLoop()
	return nil
}

func (r *PortAudioRecorder) readLoop() {
	for {
		r.mu.Lock()
		if !r.recording {
			r.mu.Unlock()
			return
		}
		stream := r.stream
		r.mu.Unlock()

		if err := stream.Read(); err != nil {
			return
		}

		r.mu.Lock()
		chunk := make([]int16, len(r.buffer))
		copy(chunk, r.buffer)
		r.samples = append(r.samples, chunk...)
		r.mu.Unlock()
	}
}

func (r *PortAudioRecorder) Stop() ([]int16, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.recording {
		return nil, nil
	}
	r.recording = false
	time.Sleep(50 * time.Millisecond)
	if err := r.stream.Stop(); err != nil {
		return nil, err
	}
	if err := r.stream.Close(); err != nil {
		return nil, err
	}
	samples := r.samples
	r.samples = nil
	return samples, nil
}

func (r *PortAudioRecorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

func Init() error {
	if err := portaudio.Initialize(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func Terminate() {
	portaudio.Terminate()
}
