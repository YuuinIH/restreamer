package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"
)

//转播实例
type streamer struct {
	mu          sync.Mutex
	Name        string        `json:"name"`
	State       streamerstate `json:"state"`
	streamurl   *url.URL
	sourceurl   *url.URL
	Autorestart bool `json:"autorestart"`
	restreamer  *ffmpeg
	waitingonce *sync.Once
}

type Streamer interface {
	GetSourceurl() *url.URL
	GetStreamurl() *url.URL
	IsAutorestart() bool
	GetName() string
	GetState() streamerstate

	StartStream() error
	StopStream() error
	RestartStream() error
	SetAutoRestart(bool) error
	UpdateStreamurl(url *url.URL) error
	UpdateSourceurl(url *url.URL) error
}

type streamerstate int

const (
	RUNNING = iota
	WAITING
	PAUSE
)

func NewStreamer(Name string, Sourceurl *url.URL, Streamurl *url.URL, Autorestart bool) (Streamer, error) {
	s := &streamer{
		Name:        Name,
		streamurl:   Streamurl,
		sourceurl:   Sourceurl,
		Autorestart: Autorestart,
		State:       PAUSE,
		waitingonce: new(sync.Once),
	}
	var err error
	if err != nil {
		return nil, err
	}
	return s, err
}

func (s *streamer) GetName() string {
	return s.Name
}

func (s *streamer) GetSourceurl() *url.URL {
	return s.sourceurl
}

func (s *streamer) GetStreamurl() *url.URL {
	return s.streamurl
}

func (s *streamer) GetState() streamerstate {
	return s.State
}

func (s *streamer) IsAutorestart() bool {
	return s.Autorestart
}

func (s *streamer) MarshalJSON() ([]byte, error) {
	type Alias streamer
	return json.Marshal(&struct {
		Streamurl string `json:"streamurl"`
		Sourceurl string `json:"sourceurl"`
		Status    int    `json:"status"`
		*Alias
	}{
		Streamurl: s.streamurl.String(),
		Sourceurl: s.sourceurl.String(),
		Alias:     (*Alias)(s),
		Status:    int(s.State),
	})
}

func (s *streamer) streamInstance() {
	var err error
	s.mu.Lock()
	s.restreamer, err = NewFFmpeg()
	if err != nil {
		s.StopStream()
		s.mu.Unlock()
		return
	}
	s.State = RUNNING
	s.mu.Unlock()

	s.restreamer.Stream(*s.sourceurl, *s.streamurl)
	//Before exit,check if the streamer is autorestart

	s.mu.Lock()
	s.waitingonce = new(sync.Once)
	if s.Autorestart {
		s.State = WAITING
		go func() {
			fmt.Printf("Streamer %s is waiting for restart\n", s.Name)
			time.Sleep(time.Second * 5)
			s.mu.Lock()
			if s.State == WAITING {
				go s.StartStream()
			}
			s.mu.Unlock()
		}()
		s.mu.Unlock()
		return
	}
	s.State = PAUSE
	s.mu.Unlock()
}

func (s *streamer) SetAutoRestart(Autorestart bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Autorestart = Autorestart
	return nil
}

func (s *streamer) UpdateStreamurl(url *url.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streamurl = url
	return nil
}

func (s *streamer) UpdateSourceurl(url *url.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sourceurl = url
	return nil
}

func (s *streamer) StopStream() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Println("Recive stop,stop streamer...")
	if s.restreamer == nil {
		return errors.New("ffmpeg is exit but state is still \"running\"")
	}
	if err := s.restreamer.Stop(); err != nil {
		return err
	}
	s.State = PAUSE
	return nil
}

func (s *streamer) StartStream() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitingonce.Do(func() {
		fmt.Println("Start streaming...")
		go s.streamInstance()
	})
	return nil
}

func (s *streamer) RestartStream() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.State {
	case PAUSE:
		s.State = WAITING
		go s.StartStream()
	case WAITING:
		go s.StartStream()
	case RUNNING:
		s.mu.Unlock()
		s.StopStream()
		s.StopStream()
		s.mu.Lock()
	}
	return nil
}
