package main

import (
	"bytes"
	"fmt"
)

func readM3U8(content []byte, out chan string) error {
	lines := bytes.Split(content, []byte("\n"))
	var first []byte

	first, lines = lines[0], lines[1:]
	fmt.Printf("first %q\n", first)
	if !bytes.Equal(first, []byte("#EXTM3U")) {
		return fmt.Errorf(".m3u8 does not start with proper header")
	}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			// meta data
			fmt.Printf("skipping %q\n", line)
			continue
		}
		if bytes.Equal(line, []byte("#EXT-X-ENDLIST")) {
			close(out)
		}
		fmt.Printf("download this %s\n", line)
		out <- fmt.Sprintf("%s", line)

	}
	return nil
}
