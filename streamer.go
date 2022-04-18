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
type Streamer struct {
	Name          string       `json:"name"`
	Status        streamstatus `json:"status"`
	Streamurl     *url.URL     `json:"streamurl,string"`
	Sourceurl     *url.URL     `json:"sourceurl,string"`
	Autorestart   bool         `json:"autorestart"`
	restreamer    *ffmpeg
	canclewaiting bool
	waitingonce   *sync.Once
}

type streamstatus interface {
	SetStreamState(streamer *Streamer)
	Startstream() error
	Stopstream() error
	Restartstream() error
	Updatestreamurl(url *url.URL) error
	Updatesourceurl(url *url.URL) error
}

type streamerstate int

const (
	RUNNING = iota
	WAITING
	PAUSE
)

func NewStreamer(Name string, Sourceurl *url.URL, Streamurl *url.URL, Autorestart bool) (*Streamer, error) {
	s := &Streamer{
		Name:        Name,
		Streamurl:   Streamurl,
		Sourceurl:   Sourceurl,
		Autorestart: Autorestart,
		Status:      &streamePause{},
	}
	var err error
	if err != nil {
		return nil, err
	}
	s.Status.SetStreamState(s)
	return s, err
}

func (s *Streamer) MarshalJSON() ([]byte, error) {
	type Alias Streamer
	return json.Marshal(&struct {
		Streamurl string `json:"streamurl"`
		Sourceurl string `json:"sourceurl"`
		Status    int    `json:"status"`
		*Alias
	}{
		Streamurl: s.Streamurl.String(),
		Sourceurl: s.Sourceurl.String(),
		Status:    s.Checkstate(),
		Alias:     (*Alias)(s),
	})
}

func (s *Streamer) Checkstate() int {
	switch s.Status.(type) {
	case *streameRunning:
		return RUNNING
	case *streameWaiting:
		return WAITING
	case *streamePause:
		return PAUSE
	}
	return -1
}

func (s *Streamer) streaminstance() {
	var err error
	s.Status = &streameRunning{
		streamer: s,
	}
	s.restreamer, err = NewFFmpeg()
	if err != nil {
		s.Stopstream()
		return
	}
	err = s.restreamer.Stream(*s.Sourceurl, *s.Streamurl)
	if s.Autorestart {
		s.Status = &streameWaiting{
			streamer: s,
		}
		s.waitingonce = new(sync.Once)
		go s.Status.Startstream()
		return
	}
	s.Status = &streamePause{
		streamer: s,
	}
	return
}

func (s *Streamer) Updatestreamurl(url *url.URL) error {
	return s.Status.Updatestreamurl(url)
}

func (s *Streamer) Updatesourceurl(url *url.URL) error {
	return s.Status.Updatesourceurl(url)
}

func (s *Streamer) Stopstream() error {
	fmt.Println("Trying stop stream...")
	return s.Status.Stopstream()
}

func (s *Streamer) Startstream() error {
	return s.Status.Startstream()
}

func (s *Streamer) Restartstream() error {
	return s.Status.Restartstream()
}

//Running
type streameRunning struct {
	streamer *Streamer
}

func (s *streameRunning) SetStreamState(streamer *Streamer) {
	s.streamer = streamer
}

func (s *streameRunning) Startstream() error {
	return errors.New("stream has been run")
}

func (s *streameRunning) Stopstream() error {
	fmt.Println("Recive stop,stop streamer...")
	if s.streamer.restreamer == nil {
		return errors.New("ffmpeg is exit but state is still \"running\"")
	}
	if err := s.streamer.restreamer.Stop(); err != nil {
		return err
	}
	s.streamer.Status = &streamePause{
		streamer: s.streamer,
	}
	return nil
}

func (s *streameRunning) Restartstream() error {
	s.streamer.Stopstream()
	s.streamer.Startstream()
	return nil
}

func (s *streameRunning) Updatesourceurl(url *url.URL) error {
	s.streamer.Sourceurl = url
	s.Restartstream()
	return nil
}

func (s *streameRunning) Updatestreamurl(url *url.URL) error {
	s.streamer.Streamurl = url
	s.Restartstream()
	return nil
}

//Waiting
type streameWaiting struct {
	streamer *Streamer
}

func (s *streameWaiting) SetStreamState(streamer *Streamer) {
	s.streamer = streamer
}

func (s *streameWaiting) Startstream() error {
	s.streamer.canclewaiting = false
	fmt.Println("Waiting...")
	time.Sleep(5 * time.Second)
	go s.streamer.waitingonce.Do(func() {
		if !s.streamer.canclewaiting {
			fmt.Println("Restart streaming...")
			go s.streamer.streaminstance()
		} else {
			s.streamer.Status = &streamePause{
				streamer: s.streamer,
			}
		}
	})
	return nil
}

func (s *streameWaiting) Stopstream() error {
	s.streamer.canclewaiting = true
	return nil
}

func (s *streameWaiting) Restartstream() error {
	return s.Startstream()
}

func (s *streameWaiting) Updatesourceurl(url *url.URL) error {
	s.streamer.Sourceurl = url
	return nil
}

func (s *streameWaiting) Updatestreamurl(url *url.URL) error {
	s.streamer.Streamurl = url
	return nil
}

//Pause
type streamePause struct {
	streamer *Streamer
}

func (s *streamePause) SetStreamState(streamer *Streamer) {
	s.streamer = streamer
}

func (s *streamePause) Startstream() error {
	fmt.Println("Start streaming...")
	go s.streamer.streaminstance()
	return nil
}

func (s *streamePause) Stopstream() error {
	return errors.New("stream has been stop")
}

func (s *streamePause) Restartstream() error {
	return s.Startstream()
}

func (s *streamePause) Updatesourceurl(url *url.URL) error {
	s.streamer.Sourceurl = url
	return nil
}

func (s *streamePause) Updatestreamurl(url *url.URL) error {
	s.streamer.Streamurl = url
	return nil
}
