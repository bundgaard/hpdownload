package main

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

func TestM3U8(t *testing.T) {
	m3u8Channel := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		link := <-m3u8Channel
		if link == "" {
			t.Errorf("expected a string. got %q", link)
		}
		wg.Done()
	}()

	fd, err := os.Open("master.m3u8")
	if err != nil {
		t.Fail()
	}

	content, _ := ioutil.ReadAll(fd)
	if err := readM3U8(content, m3u8Channel); err != nil {
		t.Errorf("%v", err)
	}

	close(m3u8Channel)
	wg.Wait()
}
