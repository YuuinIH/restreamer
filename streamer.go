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
	mu            sync.Mutex
	Name          string        `json:"name"`
	State         streamerstate `json:"state"`
	Streamurl     *url.URL      `json:"streamurl,string"`
	Sourceurl     *url.URL      `json:"sourceurl,string"`
	Autorestart   bool          `json:"autorestart"`
	restreamer    *ffmpeg
	canclewaiting bool
	waitingonce   *sync.Once
	sign          chan int
}

type Streamer interface {
	GetSourceurl() *url.URL
	GetStreamurl() *url.URL
	IsAutorestart() bool
	GetName() string
	GetState() streamerstate

	Startstream() error
	Stopstream() error
	Restartstream() error
	SetAutorestart(bool) error
	Updatestreamurl(url *url.URL) error
	Updatesourceurl(url *url.URL) error
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
		Streamurl:   Streamurl,
		Sourceurl:   Sourceurl,
		Autorestart: Autorestart,
		sign:        make(chan int),
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
	return s.Sourceurl
}

func (s *streamer) GetStreamurl() *url.URL {
	return s.Streamurl
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
		Streamurl: s.Streamurl.String(),
		Sourceurl: s.Sourceurl.String(),
		Alias:     (*Alias)(s),
	})
}

func (s *streamer) streaminstance() {
	var err error
	s.restreamer, err = NewFFmpeg()
	if err != nil {
		s.Stopstream()
		return
	}
	s.mu.Lock()
	s.State = RUNNING
	s.mu.Unlock()

	s.restreamer.Stream(*s.Sourceurl, *s.Streamurl)
	//Before exit,check if the streamer is autorestart

	s.mu.Lock()
	s.waitingonce = new(sync.Once)
	if s.Autorestart {
		s.State = WAITING
		s.mu.Unlock()
		fmt.Printf("Streamer %s is waiting for restart\n", s.Name)
		time.Sleep(time.Second * 5)
		go s.Startstream()
		return
	}
	s.State = PAUSE
	s.mu.Unlock()
}

func (s *streamer) SetAutorestart(Autorestart bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Autorestart = Autorestart
	return nil
}

func (s *streamer) Updatestreamurl(url *url.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Streamurl = url
	return nil
}

func (s *streamer) Updatesourceurl(url *url.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sourceurl = url
	return nil
}

func (s *streamer) Stopstream() error {
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

func (s *streamer) Startstream() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitingonce.Do(func() {
		fmt.Println("Restart streaming...")
		go s.streaminstance()
	})
	return nil
}

func (s *streamer) Restartstream() error {
	return nil
}
