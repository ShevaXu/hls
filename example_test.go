package hls_test

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ShevaXu/hls"
)

// Create new media playlist
// Add two segments to media playlist
// Print it
func ExampleMediaPlaylist() {
	p, _ := hls.NewMediaPlaylist(0, 2)
	p.Append(hls.QuickSegment("test01.ts", "", 5.0))
	p.Append(hls.QuickSegment("test02.ts", "", 6.0))
	fmt.Printf("%s\n", p)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:6
	// #EXTINF:5.000,
	// test01.ts
	// #EXTINF:6.000,
	// test02.ts
}

func ExampleMediaPlaylist_DecodeFrom() {
	f, _ := os.Open("sample-playlists/media-playlist-with-oatcls-scte35.m3u8")
	p, _, _ := hls.DecodeFrom(bufio.NewReader(f), true)
	pp := p.(*hls.MediaPlaylist)
	fmt.Print(pp)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:10
	// #EXT-OATCLS-SCTE35:/DAlAAAAAAAAAP/wFAUAAAABf+/+ANgNkv4AFJlwAAEBAQAA5xULLA==
	// #EXT-X-CUE-OUT:15
	// #EXTINF:8.844,
	// media0.ts
	// #EXT-X-CUE-OUT-CONT:ElapsedTime=8.844,Duration=15,SCTE35=/DAlAAAAAAAAAP/wFAUAAAABf+/+ANgNkv4AFJlwAAEBAQAA5xULLA==
	// #EXTINF:6.156,
	// media1.ts
	// #EXT-X-CUE-IN
	// #EXTINF:3.844,
	// media2.ts
}

// Example of parsing a playlist with EXT-X-DISCONTINIUTY tag
// and output it with integer segment durations.
func ExampleMediaPlaylist_DurationAsInt() {
	f, _ := os.Open("sample-playlists/media-playlist-with-discontinuity.m3u8")
	p, _, _ := hls.DecodeFrom(bufio.NewReader(f), true)
	pp := p.(*hls.MediaPlaylist)
	pp.DurationAsInt(true)
	fmt.Printf("%s", pp)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:10
	// #EXTINF:10,
	// ad0.ts
	// #EXTINF:8,
	// ad1.ts
	// #EXT-X-DISCONTINUITY
	// #EXTINF:10,
	// movieA.ts
	// #EXTINF:10,
	// movieB.ts
}

// Create new media playlist
// Add two segments to media playlist
// Print it
func ExampleMediaPlaylist_Close() {
	p, _ := hls.NewMediaPlaylist(0, 2)
	p.Append(hls.QuickSegment("test01.ts", "", 5.0))
	p.Append(hls.QuickSegment("test02.ts", "", 6.0))
	p.Close()
	fmt.Printf("%s\n", p)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-MEDIA-SEQUENCE:0
	// #EXT-X-TARGETDURATION:6
	// #EXTINF:5.000,
	// test01.ts
	// #EXTINF:6.000,
	// test02.ts
	// #EXT-X-ENDLIST
}

// Create new master playlist
// Add media playlist
// Encode structures to HLS
func ExampleMasterPlaylist() {
	m := hls.NewMasterPlaylist()
	p, _ := hls.NewMediaPlaylist(3, 5)
	for i := 0; i < 5; i++ {
		p.Append(hls.QuickSegment(fmt.Sprintf("test%d.ts", i), "", 5.0))
	}
	m.Append("chunklist1.m3u8", p, hls.VariantParams{ProgramID: 123, Bandwidth: 1500000, Resolution: "576x480"})
	m.Append("chunklist2.m3u8", p, hls.VariantParams{ProgramID: 123, Bandwidth: 1500000, Resolution: "576x480"})
	fmt.Printf("%s", m)
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:3
	// #EXT-X-STREAM-INF:PROGRAM-ID=123,BANDWIDTH=1500000,RESOLUTION=576x480
	// chunklist1.m3u8
	// #EXT-X-STREAM-INF:PROGRAM-ID=123,BANDWIDTH=1500000,RESOLUTION=576x480
	// chunklist2.m3u8
}
