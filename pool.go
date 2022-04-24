package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
)

type Streamermap map[string]*Streamer

var streamerpool Streamermap

func init() {
	streamerpool = make(map[string]*Streamer)
	if _, err := os.Stat("./data/stream.json"); os.IsNotExist(err) {
		fmt.Println("文件不存在,正在创建")
		os.Create("./stream.json")
		streamerpool.WriteFile()
		return
	}
	o, err := ioutil.ReadFile("./stream.json")
	if err != nil {
		log.Panic(err)
		return
	}
	s := []*Streamconfig{}
	err = json.Unmarshal(o, &s)
	if err != nil {
		log.Panic(err)
		return
	}
	for _, value := range s {
		err := streamerpool.Createstreamer(value)
		if err != nil {
			log.Panic(err)
			return
		}
	}
}

type Streamconfig struct {
	Name        string `json:"name"`
	Streamurl   string `json:"streamurl"`
	Sourceurl   string `json:"sourceurl"`
	Autorestart bool   `json:"autorestart"`
}

func (m Streamermap) Createstreamer(config *Streamconfig) error {
	sourceurl, err := url.Parse(config.Sourceurl)
	if err != nil {
		return err
	}
	streamurl, err := url.Parse(config.Streamurl)
	if err != nil {
		return err
	}
	s, e := m[config.Name]
	if e {
		if streamurl != s.Streamurl {
			s.Updatestreamurl(streamurl)
		}
		if sourceurl != s.Sourceurl {
			s.Updatesourceurl(sourceurl)
		}
		s.Autorestart = config.Autorestart
	} else {
		m[config.Name], err = NewStreamer(config.Name, sourceurl, streamurl, config.Autorestart)
		if err != nil {
			return err
		}
	}
	go m.WriteFile()
	return nil
}

func (m Streamermap) DeleteStreamer(name string) error {
	s, e := m[name]
	if !e {
		return errors.New("name no exited")
	}
	if s.Checkstate() != PAUSE {
		err := s.Stopstream()
		if err != nil {
			return err
		}
	}
	delete(m, name)
	go m.WriteFile()
	return nil
}

func (m Streamermap) WriteFile() {
	s := []*Streamconfig{}
	for _, value := range m {
		d := &Streamconfig{
			Name:        value.Name,
			Streamurl:   value.Streamurl.String(),
			Sourceurl:   value.Sourceurl.String(),
			Autorestart: value.Autorestart,
		}
		s = append(s, d)
	}
	j, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = ioutil.WriteFile("./data/stream.json", j, 0666)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
